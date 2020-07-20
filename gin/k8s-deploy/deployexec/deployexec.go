package deployexec

import (
	"fmt"
	"github.com/gin-gonic/gin"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

// logger 日志变量
var logger = log.New(os.Stdout, "[SSH]", log.Lshortfile|log.Ldate|log.Ltime)

var HostResultSlice []*DeploymentConfig
// DeploymentConfig 从index.html获取的配置信息
type DeploymentConfig struct {
	KubeConfig string
	Image string
	Command string
	Deployment string
	Replicas int32
	Namespace string
	DeploymentsClient v1.DeploymentInterface
	Wg sync.WaitGroup
}

// NewDeploymentConfig 构造方法
func NewDeploymentConfig(KubeConfig string,Image string,Command string,Deployment string,Replicas int32,Namespace string) *DeploymentConfig{
	KubeConfigPath := "config/"+ KubeConfig
	deploymentConfig := &DeploymentConfig{
		KubeConfig: KubeConfigPath,
		Image: Image,
		Command: Command,
		Deployment: Deployment,
		Replicas: Replicas,
		Namespace: Namespace,
	}
	//初始化deploymentsClient 字段
	deploymentConfig.clinetConfig()
	return deploymentConfig
}

func (dc *DeploymentConfig)clinetConfig() {
	config, err := clientcmd.BuildConfigFromFlags("", dc.KubeConfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	deploymentsClient := clientset.AppsV1().Deployments(dc.Namespace)
	dc.DeploymentsClient = deploymentsClient
}

func (dc *DeploymentConfig)CreateDeployment(){
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: dc.Deployment,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: dc.int32Ptr(dc.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: dc.Image,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Command: []string{"sh","-c","sleep 3600"},
						},
					},
				},
			},
		},
	}
	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := dc.DeploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
}
func (dc *DeploymentConfig)ListDeployment(){
	fmt.Printf("Listing deployments in namespace %q:\n", dc.Namespace)
	list, err := dc.DeploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}
}
func (dc *DeploymentConfig)UpdateDeployment(){
	fmt.Println("Updating deployment...")
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := dc.DeploymentsClient.Get(dc.Deployment, metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
		}

		result.Spec.Replicas = dc.int32Ptr(dc.Replicas)                           // reduce replica count
		result.Spec.Template.Spec.Containers[0].Image = dc.Image // change nginx version
		_, updateErr := dc.DeploymentsClient.Update(result)
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
	fmt.Println("Updated deployment...")
}
func (dc *DeploymentConfig)DeleteDeployment(){
	fmt.Println("Deleting deployment...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := dc.DeploymentsClient.Delete(dc.Deployment, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted deployment.")
}

func (dc *DeploymentConfig)int32Ptr(i int32) *int32 { return &i }

func (dc *DeploymentConfig)Run(){
	//dc.ListDeployment()
	dc.CreateDeployment()
}

func ExecComm(c *gin.Context){
	kubeconfig := c.PostForm("kubeconfig")
	namespace := c.PostForm("namespace")
	deployment := c.PostForm("deployment")
	command := c.PostForm("command")
	image := c.PostForm("image")
	replicas,_ := strconv.Atoi(c.PostForm("replicas"))
	//MTInstance := c.PostForm("selectInstance")
	deploymentConfig := NewDeploymentConfig(kubeconfig,image,command,deployment, int32(replicas),namespace)
	fmt.Println(deploymentConfig)
	logger.Println(deploymentConfig)
	HostResultSlice = make([]*DeploymentConfig,0)
	HostResultSlice = append(HostResultSlice,deploymentConfig)
	deploymentConfig.Run()
	c.HTML(http.StatusOK,"posts/base.html",gin.H{
		"HostResultSlice":HostResultSlice,
	})
}

func ListNameSpace(c *gin.Context){
	kubeconfig := c.PostForm("kubeconfig")
	namespace := c.PostForm("namespace")
	deployment := c.PostForm("deployment")
	command := c.PostForm("command")
	image := c.PostForm("image")
	replicas,_ := strconv.Atoi(c.PostForm("replicas"))
	//MTInstance := c.PostForm("selectInstance")
	deploymentConfig := NewDeploymentConfig(kubeconfig,image,command,deployment, int32(replicas),namespace)
	fmt.Println(deploymentConfig)
	logger.Println(deploymentConfig)
	HostResultSlice = make([]*DeploymentConfig,0)
	HostResultSlice = append(HostResultSlice,deploymentConfig)
	deploymentConfig.Run()
	c.HTML(http.StatusOK,"system/signup.html",gin.H{
		"HostResultSlice":HostResultSlice,
	})
}