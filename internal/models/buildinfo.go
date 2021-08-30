package models

type BuildInfo struct {
	Version string `json:"Version"`
	Commit  string `json:"Commit"`
	Date    string `json:"Date"`
}
