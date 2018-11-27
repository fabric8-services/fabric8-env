package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"time"

	"github.com/fabric8-services/fabric8-common/auth"
	"github.com/fabric8-services/fabric8-common/closeable"
	"github.com/fabric8-services/fabric8-common/convert/ptr"
	"github.com/fabric8-services/fabric8-common/goamiddleware"
	"github.com/fabric8-services/fabric8-common/log"
	"github.com/fabric8-services/fabric8-common/metric"
	"github.com/fabric8-services/fabric8-common/sentry"
	"github.com/fabric8-services/fabric8-env/app"
	"github.com/fabric8-services/fabric8-env/application"
	"github.com/fabric8-services/fabric8-env/configuration"
	"github.com/fabric8-services/fabric8-env/controller"
	"github.com/fabric8-services/fabric8-env/gormapp"
	"github.com/fabric8-services/fabric8-env/migration"
	"github.com/goadesign/goa"
	goalogrus "github.com/goadesign/goa/logging/logrus"
	"github.com/goadesign/goa/middleware"
	"github.com/goadesign/goa/middleware/gzip"
	"github.com/goadesign/goa/middleware/security/jwt"
	"github.com/google/gops/agent"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Parse flags
	var configFilePath string
	var printConfig bool
	var migrateDB bool
	flag.StringVar(&configFilePath, "config", "", "Path to the config file to read")
	flag.BoolVar(&printConfig, "printConfig", false, "Prints the config (including merged environment variables) and exits")
	flag.BoolVar(&migrateDB, "migrateDatabase", false, "Migrates the database to the newest version and exits.")
	flag.Parse()

	// Load config
	configFilePath = configFileFromFlags("config", "F8_CONFIG_FILE_PATH")
	config, err := configuration.New(configFilePath)
	if err != nil {
		log.Panic(nil, map[string]interface{}{"config_file_path": configFilePath, "err": err},
			"failed to setup the configuration")
	}
	if printConfig {
		os.Exit(0)
	}

	// Initialized developer mode flag and log level for the logger
	log.InitializeLogger(config.IsLogJSON(), config.GetLogLevel())

	haltSentry := initSentryClient(config)
	defer haltSentry()

	printUserInfo()

	// Init DB
	var db *gorm.DB
	db = openDB(config)
	if db != nil {
		log.Logger().Infof("DB is ok, %v", db)
		defer closeable.Close(context.Background(), db)
	}
	setupDB(db, config)

	// Migrate the schema
	err = migration.Migrate(db.DB(), config.GetPostgresDatabase())
	if err != nil {
		log.Panic(nil, map[string]interface{}{"err": err},
			"failed migration")
	}
	if migrateDB {
		os.Exit(0)
	}

	// Create service
	service := goa.New("fabric8-env")

	// Mount middlewares
	service.Use(middleware.RequestID())
	service.Use(gzip.Middleware(9))
	service.Use(app.ErrorHandler(service, true))
	service.Use(middleware.Recover())

	service.WithLogger(goalogrus.New(log.Logger()))

	tokenMgr := getTokenManager(config)
	tokenCtxMW := goamiddleware.TokenContext(tokenMgr, app.NewJWTSecurity())
	service.Use(tokenCtxMW)
	service.Use(auth.InjectTokenManager(tokenMgr))

	service.Use(log.LogRequest(config.DeveloperModeEnabled()))
	app.UseJWTMiddleware(service, jwt.New(tokenMgr.PublicKeys(), nil, app.NewJWTSecurity()))
	service.Use(metric.Recorder("fabric8_env"))
	// ---

	authService, err := auth.NewAuthService(config.GetAuthServiceURL())
	if err != nil {
		log.Panic(nil, map[string]interface{}{"url": config.GetAuthServiceURL(), "err": err},
			"could not create Auth client")
	}

	appDB := gormapp.NewGormDB(db)

	// Mount controllers
	app.MountStatusController(service, controller.NewStatusController(service, controller.NewGormDBChecker(db)))
	app.MountEnvironmentController(service, controller.NewEnvironmentController(service, appDB, authService))
	// ---

	log.Logger().Infoln("Git Commit SHA: ", app.Commit)
	log.Logger().Infoln("UTC Build Time: ", app.BuildTime)
	log.Logger().Infoln("UTC Start Time: ", app.StartTime)
	log.Logger().Infoln("GOMAXPROCS:     ", runtime.GOMAXPROCS(-1))
	log.Logger().Infoln("NumCPU:         ", runtime.NumCPU())

	http.Handle("/api/", service.Mux)
	http.Handle("/favicon.ico", http.NotFoundHandler())

	if config.GetDiagnoseHTTPAddress() != "" {
		startDiagnose(config.GetDiagnoseHTTPAddress())
	}

	registerMetrics(config, service)

	// Start http
	if err := http.ListenAndServe(config.GetHTTPAddress(), nil); err != nil {
		log.Error(nil, map[string]interface{}{"addr": config.GetHTTPAddress(), "err": err},
			"unable to connect to server")
		service.LogError("startup", "err", err)
	}
}

func configFileFromFlags(flagName string, envVarName string) string {
	configSwitchIsSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == flagName {
			configSwitchIsSet = true
		}
	})
	if !configSwitchIsSet {
		if envConfigPath, ok := os.LookupEnv(envVarName); ok {
			return envConfigPath
		}
	}
	return ""
}

func initSentryClient(config *configuration.Registry) func() {
	haltSentry, err := sentry.InitializeSentryClient(
		ptr.String(config.GetSentryDSN()),
		sentry.WithRelease(app.Commit),
		sentry.WithEnvironment(config.GetEnvironment()),
	)
	if err != nil {
		log.Panic(nil, map[string]interface{}{"err": err},
			"failed to setup the sentry client")
	}
	return haltSentry
}

func printUserInfo() {
	u, err := user.Current()
	if err != nil {
		log.Warn(nil, map[string]interface{}{"err": err},
			"failed to get current user")
	} else {
		log.Info(nil, map[string]interface{}{"username": u.Username, "uuid": u.Uid},
			"Running as user name '%s' with UID %s.", u.Username, u.Uid)
		g, err := user.LookupGroupId(u.Gid)
		if err != nil {
			log.Warn(nil, map[string]interface{}{"err": err},
				"failed to lookup group")
		} else {
			log.Info(nil, map[string]interface{}{"groupname": g.Name, "gid": g.Gid},
				"Running as as group '%s' with GID %s.", g.Name, g.Gid)
		}
	}

}

func openDB(config *configuration.Registry) *gorm.DB {
	for {
		db, err := gorm.Open("postgres", config.GetPostgresConfigString())
		if err != nil {
			log.Logger().Errorf("ERROR: Unable to open connection to database %v", err)
			log.Logger().Infof("Retrying to connect in %v...", config.GetPostgresConnectionRetrySleep())
			time.Sleep(config.GetPostgresConnectionRetrySleep())
		} else {
			return db
		}
	}
}

func setupDB(db *gorm.DB, config *configuration.Registry) {
	if config.IsPostgresDeveloperModeEnabled() && log.IsDebug() {
		db = db.Debug()
		log.Logger().Infof("Started DB debug mode")
	}
	if config.GetPostgresConnectionMaxIdle() > 0 {
		db.DB().SetMaxIdleConns(config.GetPostgresConnectionMaxIdle())
		log.Logger().Infof("Configured connection pool max idle %v", config.GetPostgresConnectionMaxIdle())
	}
	if config.GetPostgresConnectionMaxOpen() > 0 {
		db.DB().SetMaxOpenConns(config.GetPostgresConnectionMaxOpen())
		log.Logger().Infof("Configured connection pool max open %v", config.GetPostgresConnectionMaxOpen())
	}
	application.SetDatabaseTransactionTimeout(config.GetPostgresTransactionTimeout())
}

func getTokenManager(config *configuration.Registry) auth.Manager {
	tokenMgr, err := auth.DefaultManager(config)
	if err != nil {
		log.Panic(nil, map[string]interface{}{"err": err},
			"failed to setup jwt middleware")
	}
	return tokenMgr
}

func startDiagnose(addr string) {
	log.Logger().Infoln("Diagnose:       ", addr)
	if err := agent.Listen(agent.Options{Addr: addr, ConfigDir: "/tmp/gops/"}); err != nil {
		log.Error(nil, map[string]interface{}{"addr": addr, "err": err},
			"unable to connect to diagnose server")
	}
}

func registerMetrics(config *configuration.Registry, service *goa.Service) {
	if config.GetHTTPAddress() == config.GetMetricsHTTPAddress() {
		http.Handle("/metrics", promhttp.Handler())
	} else {
		go func(metricAddress string) {
			mx := http.NewServeMux()
			mx.Handle("/metrics", promhttp.Handler())
			if err := http.ListenAndServe(metricAddress, mx); err != nil {
				log.Error(nil, map[string]interface{}{
					"addr": metricAddress,
					"err":  err,
				}, "unable to connect to metrics server")
				service.LogError("startup", "err", err)
			}
		}(config.GetMetricsHTTPAddress())
	}
}
