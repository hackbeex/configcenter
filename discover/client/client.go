package client

const (
	KeyConfigClientAppIdPrefix   = "/config-client/app-id/"
	KeyConfigClientInstantPrefix = "/config-client/instance/"
	KeyConfigClientAttrCluster   = "cluster"
	KeyConfigClientAttrHost      = "host"
	KeyConfigClientAttrPost      = "post"
)

type Client struct {
	AppId   string
	Cluster string
	Host    string
	Port    int
}
