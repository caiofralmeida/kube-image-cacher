package main

import (

	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"flag"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/caiofralmeida/kube-image-cacher/handlers"
	"github.com/caiofralmeida/kube-image-cacher/internal/config"
	"github.com/caiofralmeida/kube-image-cacher/internal/registry"
	"github.com/joho/godotenv"
	"go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:path=/mutate,mutating=true,failurePolicy=ignore,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.kubeimagecacher.caiofralmeida.github.io
// +kubebuilder:rbac:groups=leases.coordination.k8s.io,resources=leases,verbs=get;list;watch

func init() {
	log.SetLogger(zap.New())
	_ = godotenv.Load()
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	entryLog := log.Log.WithName("entrypoint")

	// Setup a Manager
	entryLog.Info("setting up manager")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	/*repos, err := awsEcr.DescribeRepositories(&ecr.DescribeRepositoriesInput{
		//RegistryId: aws.String(cfg.AWSRegistryID),
	})


	fmt.Println(repos, err)
	os.Exit(0)*/

	/*creation, err := awsEcr.CreateRepository(&ecr.CreateRepositoryInput{
		RepositoryName: aws.String("foo"),
		Tags: []*ecr.Tag{
			{
				Key:   aws.String("area"),
				Value: aws.String("platform"),
			},
		},
	})

	fmt.Println(creation, err)
	os.Exit(0)*/

	ctrl.SetLogger(zap.New(
		zap.UseFlagOptions(&opts),
		zap.UseDevMode(true),
		zap.Level(zapcore.DebugLevel),
	))

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "kube-image-cacher.caiofralmeida.github.io",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		entryLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Setup application configuration
	cfg, err := config.Parse()

	if err != nil {
		entryLog.Error(err, "unable to get app config")
		os.Exit(1)
	}

	// Setup webhooks
	entryLog.Info("setting up webhook server")
	hookServer := mgr.GetWebhookServer()

	// Setup webhook handler
	awsSession := session.Must(session.NewSession())

	handler := &handlers.PodImageCacher{
		Client:     mgr.GetClient(),
		ECRService: ecr.New(awsSession),
		Registry:   registry.New(cfg.RegistryProvider, cfg.RegistryURL),
	}

	entryLog.Info("registering webhooks to the webhook server")
	hookServer.Register("/mutate", &webhook.Admission{Handler: handler})

	entryLog.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		entryLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}
