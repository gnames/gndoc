/*
Copyright © 2021 Dmitry Mozzherin <dmozzherin@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "embed"

	"github.com/gnames/gndoc"
	"github.com/gnames/gndoc/ent/doc"
	"github.com/gnames/gnfinder"
	"github.com/gnames/gnfinder/config"
	"github.com/gnames/gnfinder/ent/nlp"
	"github.com/gnames/gnfinder/ent/verifier"
	"github.com/gnames/gnfinder/io/dict"
	"github.com/gnames/gnsys"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

//go:embed gndoc.yaml
var configText string

var (
	opts []gndoc.Option
)

// cfgData purpose is to achieve automatic import of data from the
// configuration file, if it exists.
type cfgData struct {
	Format  string
	TikaURL string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gndoc [flags] docfile",
	Short: "Finds scientific names in documents and images",

	Long: `gndoc takes file name and returns back scientific names found in that
	file. The program can work with PDFs, MS Word and MS Excel documents, images
	etc. For the text extraction it uses Apache Tika service. The default service
	is located at https://tika.globalnames.org. For optional scientific name
	verification it uses gnverifier service. The default service is located at
	htts://verifier.globalnames.org.

	To see version:
	gndoc -V
	`,

	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if versionFlag(cmd) {
			os.Exit(0)
		}
		if port := portFlag(cmd); port > 0 {
			fmt.Println("...web stuff here...")
			os.Exit(0)
		}

		formatFlag(cmd)
		tikaURLFlag(cmd)
		cfg := gndoc.NewConfig(opts...)
		var data, source string
		var convDur float32
		switch len(args) {
		case 0:
			if !checkStdin() {
				_ = cmd.Help()
				os.Exit(0)
			}
			bs, err := io.ReadAll(os.Stdin)
			if err != nil {
				log.Println(err)
			}
			data = string(bs)
			source = "STDIN"
		case 1:
			source = args[0]

			d := doc.NewDoc(cfg.TikaURL)
			data, convDur, err = d.ContentFromFile(source)
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
		default:
			_ = cmd.Help()
			os.Exit(0)
		}
		findNames(data, cfg, source, convDur)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gndoc.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("version", "V", false, "Show the app version")
	rootCmd.Flags().StringP("format", "f", "", "Output format")
	rootCmd.Flags().StringP("tika_url", "t", "", "URL of Apache Tika service")
	rootCmd.Flags().IntP("port", "p", 0, "The port of gndoc web-service")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	configFile := "gndoc"
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Cannot find config directory: %s.", err)
	}

	// Search config in home directory with name ".gndoc" (without extension).
	viper.AddConfigPath(configDir)
	viper.SetConfigName(configFile)

	configPath := filepath.Join(configDir, fmt.Sprintf("%s.yaml", configFile))
	touchConfigFile(configPath, configFile)

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	getOpts()
}

// getOpts imports data from the configuration file. Some of the settings can
// be overriden by command line flags.
func getOpts() {
	cfg := &cfgData{}
	err := viper.Unmarshal(cfg)
	if err != nil {
		log.Fatalf("Cannot deserialize config data: %s.", err)
	}

	if cfg.Format != "" {
		opts = append(opts, gndoc.OptFormat(cfg.Format))
	}
	if cfg.TikaURL != "" {
		opts = append(opts, gndoc.OptTikaURL(cfg.TikaURL))
	}
}

// touchConfigFile checks if config file exists, and if not, it gets created.
func touchConfigFile(configPath string, configFile string) {
	fileExists, _ := gnsys.FileExists(configPath)
	if fileExists {
		return
	}

	log.Printf("Creating config file: %s.", configPath)
	createConfig(configPath, configFile)
}

// createConfig creates config file.
func createConfig(path string, file string) {
	err := gnsys.MakeDir(filepath.Dir(path))
	if err != nil {
		log.Fatalf("Cannot create dir %s: %s.", path, err)
	}

	err = os.WriteFile(path, []byte(configText), 0644)
	if err != nil {
		log.Fatalf("Cannot write to file %s: %s", path, err)
	}
}

func checkStdin() bool {
	stdInFile := os.Stdin
	stat, err := stdInFile.Stat()
	if err != nil {
		log.Panic(err)
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func findNames(
	data string,
	cfg gndoc.Config,
	source string,
	convDur float32,
) {
	gnfCfg := config.New()
	gnfDict := dict.LoadDictionary()
	gnfWeights := nlp.BayesWeights()
	gnf := gnfinder.New(gnfCfg, gnfDict, gnfWeights)
	res := gnf.Find(source, data)
	res.Meta.FileConversionSec = convDur

	if gnf.GetConfig().WithVerification {
		verif := verifier.New(gnf.GetConfig().PreferredSources)
		start := time.Now()
		verifiedNames := verif.Verify(res.UniqueNameStrings())
		dur := float32(time.Now().Sub(start)) / float32(time.Second)
		res.MergeVerification(verifiedNames, dur)
	}
	fmt.Println(res.Format(cfg.Format))
}
