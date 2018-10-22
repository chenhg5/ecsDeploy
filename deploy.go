package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/larspensjo/config"
)

type Project struct {
	RegionId           string
	AccessKey          string
	AccessKeySecret    string
	InstanceType       string
	HostName           string
	SecurityGroupId    string
	ImageId            string
	InstanceChargeType string
	VSwitchId          string
	Tag1Key            string
	AnsibleHostPath    string
	AnsibleHostSSHKey  string
	AnsibleHostSSHUser string
	Client             *ecs.Client
}

func NewProject(configMap map[string]string) *Project {
	return &Project{
		Client:             newClient(configMap["RegionId"], configMap["AccessKey"], configMap["AccessKeySecret"]),
		RegionId:           configMap["RegionId"],
		AccessKey:          configMap["AccessKey"],
		AccessKeySecret:    configMap["AccessKeySecret"],
		InstanceType:       configMap["InstanceType"],
		HostName:           configMap["HostName"],
		SecurityGroupId:    configMap["SecurityGroupId"],
		ImageId:            configMap["ImageId"],
		InstanceChargeType: configMap["InstanceChargeType"],
		VSwitchId:          configMap["VSwitchId"],
		Tag1Key:            configMap["Tag1Key"],
		AnsibleHostPath:    configMap["AnsibleHostPath"],
		AnsibleHostSSHKey:  configMap["AnsibleHostSSHKey"],
		AnsibleHostSSHUser: configMap["AnsibleHostSSHUser"],
	}
}

func (project *Project) UpdateAll() {
	ecsClient := project.Client

	checkReq := ecs.CreateDescribeInstancesRequest()

	var tag = []ecs.DescribeInstancesTag{{
		Key: project.Tag1Key,
		Value: time.Now().Format("20060102150405"),
	}}
	checkReq.Tag = &tag

	checkRes, err := ecsClient.DescribeInstances(checkReq)
	if err != nil {
		panic(err)
	}
	fmt.Println(checkRes)

	var Ips = make([]string, len(checkRes.Instances.Instance))

	for i := 0; i < len(checkRes.Instances.Instance); i++ {
		Ips[i] = checkRes.Instances.Instance[i].VpcAttributes.PrivateIpAddress.IpAddress[0]
	}

	fmt.Println(Ips)

	project.createHost(len(Ips), Ips)
}

func (project *Project) DeployNew(number int) {
	ecsClient := project.Client

	printTitle("创建ECS")

	createReq := ecs.CreateCreateInstanceRequest()

	var InstanceIds = `[`
	var InstanceIdsArr = make([]string, number)
	for i := 0; i < number; i++ {
		createReq.InstanceType = project.InstanceType
		createReq.HostName = project.HostName + time.Now().Format("20060102150405")
		createReq.SecurityGroupId = project.SecurityGroupId
		createReq.ImageId = project.ImageId
		createReq.RegionId = project.RegionId
		createReq.InstanceChargeType = project.InstanceChargeType
		createReq.VSwitchId = project.VSwitchId

		var tag = []ecs.CreateInstanceTag{{
			Key: project.Tag1Key,
			Value: time.Now().Format("20060102150405"),
		}}
		createReq.Tag = &tag

		response, err := ecsClient.CreateInstance(createReq)
		if err != nil {
			panic(err)
		}
		fmt.Println(response)
		InstanceIds += `"` + response.InstanceId + `"`
		InstanceIdsArr[i] = response.InstanceId
	}

	time.Sleep(time.Second * 10)

	printTitle("查询ECS")

	checkReq := ecs.CreateDescribeInstancesRequest()
	checkReq.InstanceIds = InstanceIds + `]`
	checkRes, err := ecsClient.DescribeInstances(checkReq)
	if err != nil {
		panic(err)
	}
	fmt.Println(checkRes)

	var Ips = make([]string, len(checkRes.Instances.Instance))

	for i := 0; i < len(checkRes.Instances.Instance); i++ {
		Ips[i] = checkRes.Instances.Instance[i].VpcAttributes.PrivateIpAddress.IpAddress[0]
	}

	fmt.Println(Ips)

	printTitle("启动ECS")

	//https://ecs.aliyuncs.com/?Action=StartInstance
	//&InstanceId=i-instance1

	time.Sleep(time.Second * 10)

	runReq := ecs.CreateStartInstanceRequest()

	for i := 0; i < number; i++ {
		runReq.InstanceId = InstanceIdsArr[i]
		runRes, err := ecsClient.StartInstance(runReq)
		if err != nil {
			panic(err)
		}
		fmt.Println(runRes)
	}

	project.createHost(number, Ips)
}

func (project *Project) createHost(number int, Ips []string) {
	printTitle("创建host文件")

	f, _ := os.Create(project.AnsibleHostPath)
	defer f.Close()

	title := []byte("[deploy]\n")
	f.Write(title)

	for i := 0; i < number; i++ {
		content := []byte(Ips[i] + " ansible_ssh_private_key_file=" + project.AnsibleHostSSHKey +
			" ansible_ssh_user=" + project.AnsibleHostSSHUser + "\n")
		f.Write(content)
	}

	fmt.Println("\n创建成功。")
}

func (project *Project) GetIps() []string {
	ecsClient := project.Client

	checkReq := ecs.CreateDescribeInstancesRequest()

	var tag = []ecs.DescribeInstancesTag{{
		Key:   project.Tag1Key,
		Value: time.Now().Format("20060102150405"),
	}}
	checkReq.Tag = &tag

	checkRes, err := ecsClient.DescribeInstances(checkReq)
	if err != nil {
		panic(err)
	}
	fmt.Println(checkRes)

	var Ips= make([]string, len(checkRes.Instances.Instance))

	for i := 0; i < len(checkRes.Instances.Instance); i++ {
		Ips[i] = checkRes.Instances.Instance[i].VpcAttributes.PrivateIpAddress.IpAddress[0]
	}

	return Ips
}

func newClient(regionId, accessKey, accessKeySecret string) *ecs.Client {
	ecsClient, err := ecs.NewClientWithAccessKey(
		regionId,  // 地域ID
		accessKey, // 您的Access Key ID
		accessKeySecret) // 您的Access Key Secret
	if err != nil {
		panic(err)
	}
	return ecsClient
}

func printTitle(title string) {
	fmt.Println(`// -----------------`)
	fmt.Println(`// ` + title)
	fmt.Println(`// -----------------`)
}

func GetConfig(file string, sec string) (map[string]string, error) {
	targetConfig := make(map[string]string)
	cfg, err := config.ReadDefault(file)
	if err != nil {
		return targetConfig, errors.New("unable to open config file or wrong fomart")
	}
	sections := cfg.Sections()
	if len(sections) == 0 {
		return targetConfig, errors.New("no " + sec + " config")
	}
	for _, section := range sections {
		if section != sec {
			continue
		}
		sectionData, _ := cfg.SectionOptions(section)
		for _, key := range sectionData {
			value, err := cfg.String(section, key)
			if err == nil {
				targetConfig[key] = value
			}
		}
		break
	}
	return targetConfig, nil
}
