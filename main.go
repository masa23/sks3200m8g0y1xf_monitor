package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/marpaia/graphite-golang"
	"github.com/masa23/sks3200m8g0y1xf"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Graphite struct {
		Host   string `yaml:"Host"`
		Port   int    `yaml:"Port"`
		Prefix string `yaml:"Prefix"`
	} `yaml:"Graphite"`
	Switch struct {
		Name     string `yaml:"Name"`
		URL      string `yaml:"URL"`
		User     string `yaml:"User"`
		Password string `yaml:"Password"`
	} `yaml:"Switch"`
}

func confLoad(path string) (*Config, error) {
	var conf Config

	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(buf, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func main() {
	var confPath string
	flag.StringVar(&confPath, "conf", "config.yaml", "path to config file")
	flag.Parse()

	conf, err := confLoad(confPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		now := time.Now()
		target := now.Truncate(time.Minute)
		d := target.Add(time.Minute).Sub(now)
		timer.Reset(d)
		<-timer.C

		client := sks3200m8g0y1xf.NewClient(conf.Switch.URL)
		if err := client.Login(conf.Switch.User, conf.Switch.Password); err != nil {
			log.Printf("failed to login: %v", err)
			continue
		}

		ports, err := client.GetMonitoringPortStatics()
		if err != nil {
			log.Printf("failed to get port status: %v", err)
			continue
		}

		g, err := graphite.NewGraphite(conf.Graphite.Host, conf.Graphite.Port)
		if err != nil {
			log.Printf("failed to connect to graphite: %v", err)
			continue
		}

		metrics := []graphite.Metric{}
		for _, port := range ports {
			metrics = append(metrics, graphite.Metric{
				Name:      conf.Graphite.Prefix + "." + conf.Switch.Name + "." + strconv.Itoa(port.PortNumber) + ".rx_good_pkt",
				Value:     strconv.FormatUint(port.RxGoodPkt, 10),
				Timestamp: target.Unix(),
			})
			metrics = append(metrics, graphite.Metric{
				Name:      conf.Graphite.Prefix + "." + conf.Switch.Name + "." + strconv.Itoa(port.PortNumber) + ".rx_bad_pkt",
				Value:     strconv.FormatUint(port.RxBadPkt, 10),
				Timestamp: target.Unix(),
			})
			metrics = append(metrics, graphite.Metric{
				Name:      conf.Graphite.Prefix + "." + conf.Switch.Name + "." + strconv.Itoa(port.PortNumber) + ".tx_good_pkt",
				Value:     strconv.FormatUint(port.TxGoodPkt, 10),
				Timestamp: target.Unix(),
			})
			metrics = append(metrics, graphite.Metric{
				Name:      conf.Graphite.Prefix + "." + conf.Switch.Name + "." + strconv.Itoa(port.PortNumber) + ".tx_bad_pkt",
				Value:     strconv.FormatUint(port.TxBadPkt, 10),
				Timestamp: target.Unix(),
			})
		}

		if err := g.SendMetrics(metrics); err != nil {
			log.Printf("failed to send metrics: %v", err)
			continue
		}
	}
}
