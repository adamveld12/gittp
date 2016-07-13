package gittp

import (
	"errors"
	"regexp"
)

// CreateRepo will always create a new repository if one does not exist
func CreateRepo(reponame string) bool {
	return true
}

// DenyCreateRepo will always deny creation of new repositories over git push
func DenyCreateRepo(reponame string) bool {
	return false
}

var repoRegex = regexp.MustCompile("^(?:[\\w]+)/([\\w]+)")

// UseGithubRepoNames enforces paths like /username/projectname.git
func UseGithubRepoNames(reponame string) bool {
	return repoRegex.MatchString(reponame)
}

// MasterOnly is a pre receive hook that only allows pushes to master
func MasterOnly(h *HookContext) error {
	if h.Branch == "refs/heads/master" {
		return nil
	}

	h.Fatal("Only ref updates to refs/heads/master are allowed.")

	return errors.New("hook declined")
}

// CombinePreHooks combines several PreReceiveHooks into one
func CombinePreHooks(hooks ...PreReceiveHook) PreReceiveHook {
	return func(h *HookContext) error {
		for _, prh := range hooks {
			if err := prh(h); err != nil {
				return err
			}
		}

		return nil
	}
}

// NoopPreReceive is a pre receive hook that is always successfull. This is the default if no hook is defined
func NoopPreReceive(h *HookContext) error {
	return nil
}
