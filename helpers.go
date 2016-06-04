package gittp

import "regexp"

// CreateRepo will always create a new repository if one does not exist
func CreateRepo(reponame string) bool {
	return true
}

// DenyCreateRepo will always deny creation of new repositories over git push
func DenyCreateRepo(reponame string) bool {
	return false
}

var repoRegex = regexp.MustCompile("^(?:[\\w]+)/([\\w]+).git")

// UseGithubRepoNames enforces paths like /username/projectname.git
func UseGithubRepoNames(h HookContext) bool {
	return repoRegex.MatchString(h.Repository)
}

// MasterOnly is a pre receive hook that only allows pushes to master
func MasterOnly(h HookContext) bool {
	if h.Branch == "master" {
		return true
	}

	h.Writeln("Only pushing to master is allowed.")

	return false
}

// CombinePreHooks combines several PreReceiveHooks into one
func CombinePreHooks(hooks ...PreReceiveHook) PreReceiveHook {
	return func(h HookContext) bool {
		for _, prh := range hooks {
			if !prh(h) {
				return false
			}
		}

		return true
	}
}

// NoopPreReceive is a pre receive hook that is always successfull. This is the default if no hook is defined
func NoopPreReceive(h HookContext) bool {
	return true
}
