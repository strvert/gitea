package repo

import (
	"strings"

	"code.gitea.io/gitea/modules/context"
	git_model "code.gitea.io/gitea/models/git"
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/routers/utils"
	repo_service "code.gitea.io/gitea/services/repository"
)

func NewIssueBranch(c *context.Context) {
	issue := GetActionIssue(c)
	fromBranch := c.FormString("fromBranch")
	newBranchName := c.FormString("newBranchName")

	if !c.Repo.CanCreateBranch() {
		c.NotFound("CreateBranch", nil)
		return
	}

	if strings.ContainsRune(newBranchName, ' ') {
		c.Flash.Error("ブランチ名に半角スペースが含まれています")
		c.Redirect(issue.Link())
		return
	}

	err := repo_service.CreateNewBranch(c, c.Doer, c.Repo.Repository, fromBranch, newBranchName)
	if err != nil {
		if git_model.IsErrBranchAlreadyExists(err) || git.IsErrPushOutOfDate(err) {
			c.Flash.Error(c.Tr("repo.branch.branch_already_exists", newBranchName))
			c.Redirect(issue.Link())
			return
		}
		if git_model.IsErrBranchNameConflict(err) {
			e := err.(git_model.ErrBranchNameConflict)
			c.Flash.Error(c.Tr("repo.branch.branch_name_conflict", newBranchName, e.BranchName))
			c.Redirect(issue.Link())
			return
		}
		if git.IsErrPushRejected(err) {
			e := err.(*git.ErrPushRejected)
			if len(e.Message) == 0 {
				c.Flash.Error(c.Tr("repo.editor.push_rejected_no_message"))
			} else {
				flashError, err := c.RenderToString(tplAlertDetails, map[string]interface{}{
					"Message": c.Tr("repo.editor.push_rejected"),
					"Summary": c.Tr("repo.editor.push_rejected_summary"),
					"Details": utils.SanitizeFlashErrorString(e.Message),
				})
				if err != nil {
					c.ServerError("UpdatePullRequest.HTMLString", err)
					return
				}
				c.Flash.Error(flashError)
			}
			c.ServerError("CreateNewBranch", err)
			return
		}
	}

	c.Redirect(issue.Link())
}
