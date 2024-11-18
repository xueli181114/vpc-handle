package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/openshift-online/ocm-common/pkg/aws/aws_client"
	"github.com/openshift-online/ocm-common/pkg/test/vpc_client"
)

var region string = "us-west-2"
var vpcKeyword string = "gate-stg-cs-ci" // all gating jobs prepared vpc are started from gate-<env>-cs-ci
var timeDu time.Duration = 6             // the vpc created before Now - timeDu hours will be deleted

func CleanUpVPC() {
	client, err := aws_client.CreateAWSClient("", region)
	if err != nil {
		panic(err.Error())
	}
	vpcs, err := client.ListVPCs()
	if err != nil {
		panic(err.Error())
	}
	vpcCleanup := []string{}
	throttleTime := time.Now()
	for _, vpcExist := range vpcs {
		nameMatch := false
		timeMatch := false
		for _, tag := range vpcExist.Tags {
			if *tag.Key == "Name" && (strings.Contains(*tag.Value, vpcKeyword)) {
				nameMatch = true
			}
			if *tag.Key == "openshift_creationDate" {
				creationTime, err := time.Parse(time.RFC3339, *tag.Value)
				if err != nil {
					panic(err.Error())
				}
				if creationTime.Add(timeDu * time.Hour).Before(throttleTime) {
					timeMatch = true
				}
			}
		}
		if nameMatch && timeMatch {
			fmt.Println(">>> Got one matched vpc need to clean: ", *vpcExist.VpcId)
			vpcCleanup = append(vpcCleanup, *vpcExist.VpcId)
		}
	}

	wg := sync.WaitGroup{}
	for _, vpcID := range vpcCleanup {
		wg.Add(1)
		go func(vpcID string) {
			defer wg.Done()
			vpc, err := vpc_client.GenerateVPCByID(vpcID, region)
			if err != nil {
				return
			}
			err = vpc.DeleteVPCChain(true)
			if err != nil {
				fmt.Println(">>>> vpc ID: ", vpc.VpcID)
			}
		}(vpcID)
	}

	wg.Wait()

}
func FindLauncImage() {
	vpc, _ := vpc_client.PrepareVPC("xueli", region, "", true, "")
	imageID, err := vpc.FindProxyLaunchImage()
	fmt.Println(err)
	fmt.Println(imageID)
}
func LaunchProxy() {
	vpc, _ := vpc_client.PrepareVPC("xueli", region, "", true, "")
	defer vpc.DeleteVPCChain(true)
	_, ip, ca, err := vpc.LaunchProxyInstance(region+"a", "xuelitmp2", "/Users/lixue/Workspace/ocm-common")
	fmt.Println(ip)
	fmt.Println(ca)
	fmt.Println(err)
}

func main() {
	CleanUpVPC()
}
