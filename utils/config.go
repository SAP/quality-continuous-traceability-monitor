package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/glog"
)

const envKeyGithubAccesstoken = "GITHUB_TOKEN"

// Git coordinates
type Git struct {
	Organization string
	Repository   string
	Branch       string `json:"branch"`
}

// Delivery JSON
type delivery struct {
	ProgramName     string   `json:"program"`
	ProgramDelivery string   `json:"delivery"`
	JiraKeys        []string `json:"jira_keys,omitempty"`
	GithubKeys      []string `json:"github_keys,omitempty"`
}

// Sourcecode location (local and/or remote) and coding language
type Sourcecode struct {
	Local       string
	Git         Git
	Language    string
	CustomURLTemplate string
}

// Config struct representation of your JSON config file
type Config struct {
	Github struct {
		AccessToken               string `json:"access_token,omitempty"`
		BaseURL                   string `json:"base_url"`
		CreateLinksInBacklogItems bool   `json:"createLinksInBacklogItems,omitempty"`
	} `json:"github,omitempty"`
	Jira struct {
		BaseURL   string `json:"base_url,omitempty"`
		BasicAuth struct {
			User     string `json:"user"`
			Password string `json:"password"`
		} `json:"basicAuth,omitempty"`
		CreateLinksInBacklogItems bool `json:"createLinksInBacklogItems,omitempty"`
	} `json:"jira,omitempty"`
	Sourcecode []Sourcecode
	Mapping    struct {
		Local string
	}
	TestReport []struct {
		Type  string
		Local string
	}
	TraceabilityRepo struct {
		Git Git
	}
	Delivery struct {
		Program      string
		Version      string
		Backlogitems string
	}
	WorkDir   string
	OutputDir string
	Log       struct {
		Level string
	}
}

// ReadConfig from json file
func (cfg *Config) ReadConfig(configFilePath *string) {

	if Exists(*configFilePath) == false {
		glog.Fatal("Given config file does not exists/no authorization (Given file was ", *configFilePath, ")")
	}

	dat, err := ioutil.ReadFile(*configFilePath)
	if err != nil {
		panic(err)
	}

	// Get struct data type from json file
	err = json.Unmarshal(dat, &cfg)
	if err != nil {
		glog.Fatal("Unable to read config file (", *configFilePath, "). Are you sure this is a valid JSON config file? Nested error was: ", err)
	}

	// Check if temp dir exists (required for following steps)
	if Exists(cfg.WorkDir) == false {
		err := os.MkdirAll(cfg.WorkDir, os.FileMode(0755))
		if err != nil {
			glog.Fatal("Given temporary working dir does not exist and I cannot create it. (Given temp dir was: ", cfg.WorkDir, ")")
		}
	}

	// Check if output dir exists
	if Exists(cfg.OutputDir) == false {
		err := os.MkdirAll(cfg.OutputDir, os.FileMode(0755))
		if err != nil {
			glog.Fatal("Given output dir does not exist and I cannot create it. (Given output dir was: ", cfg.OutputDir, ")")
		}
	}

	// Read configuration from environment (e.g. credentials and stuff)
	// Needs to be done, before we start cloning repos etc.
	cfg.readEnvironment()

	// We only clone the src code repo if new don't have a mapping file (-> we need to parse the source code by ourself)
	if cfg.Mapping.Local == "" {
		for x, sc := range cfg.Sourcecode {
			// Check if sourcecode repository is already cloned locally
			if sc.Local != "" { // Local path is given...
				if Exists(sc.Local) == false { // ...but does not exist
					sourcecodeRepoURL := GetRepositoryHTTPSUrl(*cfg, sc.Git)
					glog.Info("Cloning sourcecode repository ", sourcecodeRepoURL, " to ", sc.Local)
					if IsGitInstalled() == false {
						glog.Fatal("You specified a local sourcecode repository, which could not be found. So I tried to clone it, but failed\n" +
							" as you don't have a git command line client installed.\n" +
							" Please please check the local source code repository path and/or install the git command line client.")
					}
					CloneRepo(*cfg, sourcecodeRepoURL, sc.Local)
				}
			} else {
				if sc.Git.Organization == "" {
					glog.Fatal("No mapping file and no sourcecode repository given. Work so I can not!")
				}
				if IsGitInstalled() == false {
					glog.Fatal("As you specified remote sourcecode repository, but I cannot find a git command line client.\n" +
						"Please install a git command line client or (if the sourcecode is already cloned locally) set the path to the local sourcecode.")
				}
				sourcecodeRepoURL := GetRepositoryHTTPSUrl(*cfg, sc.Git)
				repoName := GetRepositoryNameFromURL(sourcecodeRepoURL)
				localPath := cfg.WorkDir + string(os.PathSeparator) + repoName
				glog.Info("Cloning sourcecode repository ", sourcecodeRepoURL, " to ", localPath)
				CloneOrPullRepo(*cfg, sourcecodeRepoURL, localPath, true)
				cfg.Sourcecode[x].Local = localPath // Set config value
			}
		}
	}

	// Check if github basedir (and access key) exist
	if cfg.Github.BaseURL != "" {
		cfg.Github.BaseURL = strings.Trim(cfg.Github.BaseURL, "/")
	} else {
		glog.Warning("No GitHub configuration given! Links in generated reports may be broken. Do not use the trceability repository. Other strange things might occur.")
	}

	// Check traceability repo and its prerequisites
	if cfg.TraceabilityRepo.Git.Repository != "" {
		// Check prerequisites
		if IsGitInstalled() == false {
			glog.Fatal("As you specified a traceability repo, you need to have a git command line client installed.")
		}

		// Commented out because if there is no access token, git will try to use SSH key (deploy key)
		// if cfg.Github.Access_token == "" {
		// 	glog.Fatal("As you specified a traceability repo, you also need to specify a GitHub access key so that continuous quality metrics can be pushed to GitHub.")
		// }
	}

	// Check if mapping file is given and if so, does it exist?
	if cfg.Mapping.Local != "" {
		if Exists(cfg.Mapping.Local) == false {
			glog.Fatal("Given mapping file does not exist (Given mapping file was: ", cfg.Mapping.Local, ")")
		}
	}

	// Check if test reports dirs exists (and that they contain files)
	for _, trPath := range cfg.TestReport {

		if Exists(trPath.Local) == false {
			glog.Fatal("Given test report directory does not exist (Given test report directory was: ", trPath.Local)
		} else {
			files, err := ioutil.ReadDir(trPath.Local)
			if err != nil {
				glog.Fatal("Unable to read test report directory (Given test report directory was: ", trPath.Local)
			}
			if len(files) == 0 {
				glog.Fatal("Test report directory is empty!? (Given test report directory was: ", trPath.Local)
			}
		}
	}

}

// ReadDelivery from json file
func (cfg *Config) ReadDelivery(deliveryFilePath *string) {

	if Exists(*deliveryFilePath) == false {
		glog.Error("Given delivery file does not exists/no authorization (Given file was ", *deliveryFilePath, ")")
		return
	}

	dat, err := ioutil.ReadFile(*deliveryFilePath)
	if err != nil {
		panic(err)
	}

	var del = delivery{}

	// Get struct data type from json file
	err = json.Unmarshal(dat, &del)
	if err != nil {
		glog.Fatal("Unable to read local delivery file (", *deliveryFilePath, "). Are you sure this is a valid JSON delivery file? Nested error was: ", err)
	}

	cfg.Delivery.Program = strings.Replace(del.ProgramName, " ", "", -1)
	cfg.Delivery.Version = strings.Replace(del.ProgramDelivery, " ", "", -1)

	var backlogItems string
	for i, ghk := range del.GithubKeys {
		if i == 0 {
			backlogItems = "GitHub:" + ghk
		} else {
			backlogItems = backlogItems + ", " + "GitHub:" + ghk
		}
	}
	for i, jk := range del.JiraKeys {
		if i == 0 && backlogItems == "" {
			backlogItems = "Jira:" + jk
		} else {
			backlogItems = backlogItems + ", " + "Jira:" + jk
		}
	}
	cfg.Delivery.Backlogitems = backlogItems

}

func (cfg *Config) readEnvironment() {

	ghAccessToken := os.Getenv(envKeyGithubAccesstoken)
	if ghAccessToken != "" {
		cfg.Github.AccessToken = ghAccessToken
	}

}

// Exists file?
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
