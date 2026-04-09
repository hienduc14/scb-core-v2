package valueobjects

type ProjectType string

const (
	ProjectTypeGitHub       ProjectType = "GITHUB"
	ProjectTypeGitHubServer ProjectType = "GITHUB_SERVER"
	ProjectTypeGitLab       ProjectType = "GITLAB"
	ProjectTypeGitLabServer ProjectType = "GITLAB_SERVER"
)

func (p ProjectType) IsSupported() bool {
	switch p {
	case ProjectTypeGitHub, ProjectTypeGitHubServer, ProjectTypeGitLab, ProjectTypeGitLabServer:
		return true
	default:
		return false
	}
}
