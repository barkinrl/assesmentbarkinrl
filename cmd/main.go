/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/kazimsarikaya/assesmentbarkinrl/internal/config"
	"github.com/kazimsarikaya/assesmentbarkinrl/internal/logger"
	"github.com/kazimsarikaya/assesmentbarkinrl/internal/webserver"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// define cobra/viper root command
var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "app",
		Short: "app is a simple app server",
		Long:  `app is a simple app server`,
	}

	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Start the app server",
		Long:  `Start the app server with the specified options`,
		Run: func(cmd *cobra.Command, args []string) {
			err := cmdServer()

			if err != nil {
				slog.Error("Error starting server", "error", err)
			}
		},
	}

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of application",
		Long:  `Print the version number of application`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("app version:", config.GetConfig().GetVersion())
			fmt.Println("build time:", config.GetConfig().GetBuildTime())
			fmt.Println("go version:", config.GetConfig().GetGoVersion())
		},
	}
)

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	config.GetConfigBuilder().SyncConfig()
}

func init() {
	slog.SetDefault(logger.DefaultSLogger)
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.app.yaml)")
	err := viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	if err != nil {
		slog.Error("Error binding config flag", "error", err)
	}

	viper.SetDefault("config", "")

	cb := config.GetConfigBuilder()

	cb.BuildCommandlineFlags(rootCmd, serverCmd)

	rootCmd.AddCommand(serverCmd)

	rootCmd.AddCommand(versionCmd)
}

func cmdServer() error {
	config := config.GetConfig()

	slog.Info("config", "server_port", config.GetServerPort())
	slog.Info("config", "debug", config.GetDebug())

	if config.GetDebug() {
		logger.LogLevel.Set(slog.LevelDebug)
	}

	srv, err := webserver.StartWebServer()

	if err != nil {
		return err
	}

	c := make(chan os.Signal, 1)

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), config.GetWait())
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	err = srv.Shutdown(ctx)

	if err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	slog.Info("Shutting down")
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("cannot execute", "command", rootCmd.Use, "error", err)
		os.Exit(1)
	}
}
