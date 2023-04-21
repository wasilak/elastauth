package libs

import (
	"flag"
	"fmt"
	"os"

	"github.com/labstack/gommon/log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func InitConfiguration() error {
	flag.Bool("debug", false, "Debug")
	flag.Bool("generateKey", false, "Generate valid encryption key for use in app")
	flag.String("listen", "127.0.0.1:5000", "Listen address")
	flag.String("config", "./", "Path to config.yml")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetEnvPrefix("elastauth")
	viper.AutomaticEnv()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(viper.GetString("config"))

	viper.SetDefault("cache_type", "memory")
	viper.SetDefault("redis_host", "localhost:6379")
	viper.SetDefault("redis_db", 0)
	viper.SetDefault("cache_expire", "1h")
	viper.SetDefault("elasticsearch_dry_run", false)

	viper.SetDefault("headers_username", "Remote-User")
	viper.SetDefault("headers_groups", "Remote-Groups")
	viper.SetDefault("headers_Email", "Remote-Email")
	viper.SetDefault("headers_name", "Remote-Name")

	viper.SetDefault("enable_metrics", false)

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
		return err
	}

	if viper.GetBool("debug") {
		log.SetLevel(log.DEBUG)
	}

	return nil
}

func HandleSecretKey() {
	if viper.GetBool("generateKey") {
		key, err := GenerateKey()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(key)
		os.Exit(0)
	}

	if len(viper.GetString("secret_key")) == 0 {
		key, err := GenerateKey()
		if err != nil {
			log.Fatal(err)
		}
		viper.Set("secret_key", key)
		log.Info(fmt.Sprintf("WARNING: No secret key provided. Setting randomly generated: %s", key))
	}
}
