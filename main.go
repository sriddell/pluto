package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/fatih/color"
	rConfig "github.com/sriddell/k8s-lib/config"
	"github.com/sriddell/k8s-lib/rancher"

	"os"
)

type config struct {
	Servers []rancher.RancherServer `json:"rancherServers"`
}

var red = color.New(color.FgRed).PrintfFunc()

func main() {

	jsonFile, err := os.Open("config.json")
	if err != nil {
		panic(err.Error())
	}
	defer jsonFile.Close()
	b, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err.Error())
	}
	var c config
	json.Unmarshal(b, &c)
	rancherServers := c.Servers
	for i := 0; i < len(rancherServers); i++ {
		clusters := rancher.GetClusters(rancherServers[i])
		pluto(rancherServers[i], clusters)
	}
}


func pluto(rancherServer rancher.RancherServer, clusters []rancher.Cluster) {
	for i := 0; i < len(clusters); i++ {
		c := clusters[i]
		if c.Name == "local" {
			continue
		}
		fmt.Printf("rancher=%v\n", rancherServer.RancherUrl)
		fmt.Printf("name=%v\n", c.Name)
		fmt.Printf("clusterId=%v\n", c.Id)
		rawConfig, err := rConfig.GetKubeConfig(c.Actions["generateKubeconfig"], rancherServer.Token)
		if err != nil {
			red(">>cluster details not available; may be updating control nodes<<\n\n")
			os.Exit(1)
		}
		f, err := os.Create("./kubeconfig")
		if err != nil {
			red("Could not create kubeconfig\n")
			os.Exit(1)
		}
		f.Write(rawConfig)
		f.Close()

		// cmd := exec.Command("pluto", "detect-helm", "-t", "k8s=v1.25.0", "--output", "markdown")
		cmd := exec.Command("pluto", "detect-helm", "-t", "k8s=v1.25.0", "-r", "--output", "markdown")
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "KUBECONFIG=./kubeconfig")
		// open the out file for writing
		outfile, err := os.Create(c.Name + ".md")
		if err != nil {
			panic(err)
		}
		defer outfile.Close()
		cmd.Stdout = outfile

		err = cmd.Start()
		if err != nil {
			panic(err)
		}
		cmd.Wait()
	}
}
