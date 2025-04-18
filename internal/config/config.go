/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
package config

import (
	"log/slog"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version   string
	buildTime string
	goVersion string
)

type ConfigBuilder interface {
	BuildCommandlineFlags(rootCmd *cobra.Command, serverCmd *cobra.Command)
	SyncConfig()
}

type Config interface {
	GetServerPort() int
	GetDebug() bool
	GetWait() time.Duration
	GetOidcIssuer() string
	GetOidcAudience() string
	GetLocalStaticPath() string
	GetKubeCAFile() string
	GetKubeApiServer() string
	GetVersion() string
	GetBuildTime() string
	GetGoVersion() string
}

type config struct {
	serverPort      int
	debug           bool
	wait            time.Duration
	rotateTimer     *time.Timer
	oidcIssuer      string
	oidcAudience    string
	localStaticPath string
	kubeCAFile      string
	kubeApiServer   string
}

var (
	_config *config = nil
)

func getConfigSingleton() *config {
	if _config == nil {

		rt := time.NewTimer(1 * time.Hour)

		go func() {
			for {
				<-rt.C

				rt.Reset(1 * time.Hour)
			}
		}()

		_config = &config{
			rotateTimer: rt,
		}
	}

	return _config
}

func GetConfig() Config {
	return getConfigSingleton()
}

func GetConfigBuilder() ConfigBuilder {
	return getConfigSingleton()
}

func (c *config) BuildCommandlineFlags(rootCmd *cobra.Command, serverCmd *cobra.Command) {
	rootCmd.PersistentFlags().BoolVarP(&c.debug, "debug", "d", false, "Enable debug mode")

	err := viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	if err != nil {
		slog.Error("Error binding debug flag", "error", err)
	}

	viper.SetDefault("debug", false)

	serverCmd.Flags().IntVarP(&c.serverPort, "serverPort", "p", 0, "Port to listen on")
	err = viper.BindPFlag("serverPort", serverCmd.Flags().Lookup("serverPort"))

	if err != nil {
		slog.Error("Error binding serverPort flag", "error", err)
	}

	viper.SetDefault("serverPort", 8080)

	serverCmd.Flags().DurationVarP(&c.wait, "wait", "w", 0, "Time to wait before shutting down")
	err = viper.BindPFlag("wait", serverCmd.Flags().Lookup("wait"))

	if err != nil {
		slog.Error("Error binding wait flag", "error", err)
	}

	viper.SetDefault("wait", 15*time.Second)

	serverCmd.Flags().StringVarP(&c.oidcIssuer, "oidcIssuer", "", "", "OIDC Issuer")
	err = viper.BindPFlag("oidcIssuer", serverCmd.Flags().Lookup("oidcIssuer"))

	if err != nil {
		slog.Error("Error binding oidcIssuer flag", "error", err)
	}

	viper.SetDefault("oidcIssuer", "")

	serverCmd.Flags().StringVarP(&c.oidcAudience, "oidcAudience", "", "", "OIDC Audience")
	err = viper.BindPFlag("oidcAudience", serverCmd.Flags().Lookup("oidcAudience"))

	if err != nil {
		slog.Error("Error binding oidcAudience flag", "error", err)
	}

	viper.SetDefault("oidcAudience", "")

	serverCmd.Flags().StringVarP(&c.localStaticPath, "localStaticPath", "", "", "Local path to static files")
	err = viper.BindPFlag("localStaticPath", serverCmd.Flags().Lookup("localStaticPath"))

	if err != nil {
		slog.Error("Error binding localStaticPath flag", "error", err)
	}

	viper.SetDefault("localStaticPath", "")

	serverCmd.Flags().StringVarP(&c.kubeCAFile, "kubeCAFile", "", "", "Kubernetes CA file")
	err = viper.BindPFlag("kubeCAFile", serverCmd.Flags().Lookup("kubeCAFile"))

	if err != nil {
		slog.Error("Error binding kubeCAFile flag", "error", err)
	}

	viper.SetDefault("kubeCAFile", "")

	serverCmd.Flags().StringVarP(&c.kubeApiServer, "kubeApiServer", "", "", "Kubernetes API server")
	err = viper.BindPFlag("kubeApiServer", serverCmd.Flags().Lookup("kubeApiServer"))

	if err != nil {
		slog.Error("Error binding kubeApiServer flag", "error", err)
	}

	viper.SetDefault("kubeApiServer", "")
}

func (c *config) SyncConfig() {
	c.debug = viper.GetBool("debug")
	c.serverPort = viper.GetInt("serverPort")
	c.wait = viper.GetDuration("wait")
	c.oidcIssuer = viper.GetString("oidcIssuer")
	c.oidcAudience = viper.GetString("oidcAudience")
	c.localStaticPath = viper.GetString("localStaticPath")
	c.kubeCAFile = viper.GetString("kubeCAFile")
	c.kubeApiServer = viper.GetString("kubeApiServer")
}

func (c *config) GetServerPort() int {
	return c.serverPort
}

func (c *config) GetDebug() bool {
	return c.debug
}

func (c *config) GetWait() time.Duration {
	return c.wait
}

func (c *config) GetOidcIssuer() string {
	return c.oidcIssuer
}

func (c *config) GetOidcAudience() string {
	return c.oidcAudience
}

func (c *config) GetLocalStaticPath() string {
	return c.localStaticPath
}

func (c *config) GetKubeCAFile() string {
	return c.kubeCAFile
}

func (c *config) GetKubeApiServer() string {
	return c.kubeApiServer
}

func (c *config) GetVersion() string {
	return version
}

func (c *config) GetBuildTime() string {
	return buildTime
}

func (c *config) GetGoVersion() string {
	return goVersion
}
