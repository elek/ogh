package main

var artifactMap = map[string]string{
	"integration (freon)":               "it-freon",
	"integration (filesystem)":          "it-filesystem",
	"integration (filesystem-contract)": "it-filesystem-contract",
	"integration (client)":              "it-client",
	"integration (hdds-om)":             "it-hdds-om",
	"integration (ozone)":               "it-ozone",
}

func JobToArtifactName(job string) string {
	if artifact, ok := artifactMap[job]; ok {
		return artifact
	}
	return job
}