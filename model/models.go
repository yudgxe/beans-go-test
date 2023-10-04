package model

type Packet struct {
	Name    string `json:"name"`
	Version string `json:"ver"`
}
type Target struct {
	Path    string `json:"path"`
	Exclude string `json:"exclude"`
}

type Create struct {
	Packet
	Targets []Target `json:"targets"`
	Packets []Packet `json:"packets"`
}

type Update struct {
	Packets []Packet `json:"packages"`
}
