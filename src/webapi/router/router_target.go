package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/common/model"
	"github.com/toolkits/pkg/ginx"

	"github.com/didi/nightingale/v5/src/models"
)

func targetGets(c *gin.Context) {
	bgid := ginx.QueryInt64(c, "bgid", -1)
	query := ginx.QueryStr(c, "query", "")
	limit := ginx.QueryInt(c, "limit", 30)
	clusters := queryClusters(c)

	total, err := models.TargetTotal(bgid, clusters, query)
	ginx.Dangerous(err)

	list, err := models.TargetGets(bgid, clusters, query, limit, ginx.Offset(c, limit))
	ginx.Dangerous(err)

	if err == nil {
		cache := make(map[int64]*models.BusiGroup)
		for i := 0; i < len(list); i++ {
			ginx.Dangerous(list[i].FillGroup(cache))
		}
	}

	ginx.NewRender(c).Data(gin.H{
		"list":  list,
		"total": total,
	}, nil)
}

func targetGetTags(c *gin.Context) {
	idents := ginx.QueryStr(c, "idents")
	idents = strings.ReplaceAll(idents, ",", " ")
	lst, err := models.TargetGetTags(strings.Fields(idents))
	ginx.NewRender(c).Data(lst, err)
}

type targetTagsForm struct {
	Idents []string `json:"idents" binding:"required"`
	Tags   []string `json:"tags" binding:"required"`
}

func (t targetTagsForm) Verify() {

}

func targetBindTagsByFE(c *gin.Context) {
	var f targetTagsForm
	ginx.BindJSON(c, &f)

	if len(f.Idents) == 0 {
		ginx.Bomb(http.StatusBadRequest, "idents empty")
	}

	checkTargetPerm(c, f.Idents)

	ginx.NewRender(c).Message(targetBindTags(f))
}

func targetBindTagsByService(c *gin.Context) {
	var f targetTagsForm
	ginx.BindJSON(c, &f)

	if len(f.Idents) == 0 {
		ginx.Bomb(http.StatusBadRequest, "idents empty")
	}

	ginx.NewRender(c).Message(targetBindTags(f))
}

func targetBindTags(f targetTagsForm) error {
	for i := 0; i < len(f.Tags); i++ {
		arr := strings.Split(f.Tags[i], "=")
		if len(arr) != 2 {
			return fmt.Errorf("invalid tag(%s)", f.Tags[i])
		}

		if strings.TrimSpace(arr[0]) == "" || strings.TrimSpace(arr[1]) == "" {
			return fmt.Errorf("invalid tag(%s)", f.Tags[i])
		}

		if strings.IndexByte(arr[0], '.') != -1 {
			return fmt.Errorf("invalid tagkey(%s): cannot contains . ", arr[0])
		}

		if strings.IndexByte(arr[0], '-') != -1 {
			return fmt.Errorf("invalid tagkey(%s): cannot contains -", arr[0])
		}

		if !model.LabelNameRE.MatchString(arr[0]) {
			return fmt.Errorf("invalid tagkey(%s)", arr[0])
		}
	}

	for i := 0; i < len(f.Idents); i++ {
		target, err := models.TargetGetByIdent(f.Idents[i])
		if err != nil {
			return err
		}

		if target == nil {
			continue
		}

		// ????????????key?????????????????????????????????????????????????????????????????????
		for j := 0; j < len(f.Tags); j++ {
			tagkey := strings.Split(f.Tags[j], "=")[0]
			tagkeyPrefix := tagkey + "="
			if strings.HasPrefix(target.Tags, tagkeyPrefix) {
				return fmt.Errorf("duplicate tagkey(%s)", tagkey)
			}
		}

		err = target.AddTags(f.Tags)
		if err != nil {
			return err
		}
	}
	return nil
}

func targetUnbindTagsByFE(c *gin.Context) {
	var f targetTagsForm
	ginx.BindJSON(c, &f)

	if len(f.Idents) == 0 {
		ginx.Bomb(http.StatusBadRequest, "idents empty")
	}

	checkTargetPerm(c, f.Idents)

	ginx.NewRender(c).Message(targetUnbindTags(f))
}

func targetUnbindTagsByService(c *gin.Context) {
	var f targetTagsForm
	ginx.BindJSON(c, &f)

	if len(f.Idents) == 0 {
		ginx.Bomb(http.StatusBadRequest, "idents empty")
	}

	ginx.NewRender(c).Message(targetUnbindTags(f))
}

func targetUnbindTags(f targetTagsForm) error {
	for i := 0; i < len(f.Idents); i++ {
		target, err := models.TargetGetByIdent(f.Idents[i])
		if err != nil {
			return err
		}

		if target == nil {
			continue
		}

		err = target.DelTags(f.Tags)
		if err != nil {
			return err
		}
	}
	return nil
}

type targetNoteForm struct {
	Idents []string `json:"idents" binding:"required"`
	Note   string   `json:"note"`
}

func targetUpdateNote(c *gin.Context) {
	var f targetNoteForm
	ginx.BindJSON(c, &f)

	if len(f.Idents) == 0 {
		ginx.Bomb(http.StatusBadRequest, "idents empty")
	}

	checkTargetPerm(c, f.Idents)

	ginx.NewRender(c).Message(models.TargetUpdateNote(f.Idents, f.Note))
}

func targetUpdateNoteByService(c *gin.Context) {
	var f targetNoteForm
	ginx.BindJSON(c, &f)

	if len(f.Idents) == 0 {
		ginx.Bomb(http.StatusBadRequest, "idents empty")
	}

	ginx.NewRender(c).Message(models.TargetUpdateNote(f.Idents, f.Note))
}

type targetBgidForm struct {
	Idents []string `json:"idents" binding:"required"`
	Bgid   int64    `json:"bgid"`
}

func targetUpdateBgid(c *gin.Context) {
	var f targetBgidForm
	ginx.BindJSON(c, &f)

	if len(f.Idents) == 0 {
		ginx.Bomb(http.StatusBadRequest, "idents empty")
	}

	user := c.MustGet("user").(*models.User)
	if user.IsAdmin() {
		ginx.NewRender(c).Message(models.TargetUpdateBgid(f.Idents, f.Bgid, false))
		return
	}

	if f.Bgid > 0 {
		// ???????????????????????????????????????????????????bgid???0???????????????????????????????????????bgid>0????????????????????????????????????
		// ?????????????????????didiyun????????????didiyun???????????????????????????????????????didiyun-ceph???
		// ?????????????????????????????????????????????????????????????????????????????????????????????????????????BG???????????????
		orphans, err := models.IdentsFilter(f.Idents, "group_id = ?", 0)
		ginx.Dangerous(err)

		// ?????????????????????????????????????????????????????????admin
		if len(orphans) > 0 && !user.IsAdmin() {
			ginx.Bomb(http.StatusForbidden, "No permission. Only admin can assign BG")
		}

		reBelongs, err := models.IdentsFilter(f.Idents, "group_id > ?", 0)
		ginx.Dangerous(err)

		if len(reBelongs) > 0 {
			// ??????????????????????????????????????????????????????????????????????????????????????????????????????bgid?????????
			checkTargetPerm(c, f.Idents)

			bg := BusiGroup(f.Bgid)
			can, err := user.CanDoBusiGroup(bg, "rw")
			ginx.Dangerous(err)

			if !can {
				ginx.Bomb(http.StatusForbidden, "No permission. You are not admin of BG(%s)", bg.Name)
			}
		}
	} else if f.Bgid == 0 {
		// ????????????
		checkTargetPerm(c, f.Idents)
	} else {
		ginx.Bomb(http.StatusBadRequest, "invalid bgid")
	}

	ginx.NewRender(c).Message(models.TargetUpdateBgid(f.Idents, f.Bgid, false))
}

type identsForm struct {
	Idents []string `json:"idents" binding:"required"`
}

func targetDel(c *gin.Context) {
	var f identsForm
	ginx.BindJSON(c, &f)

	if len(f.Idents) == 0 {
		ginx.Bomb(http.StatusBadRequest, "idents empty")
	}

	checkTargetPerm(c, f.Idents)

	ginx.NewRender(c).Message(models.TargetDel(f.Idents))
}

func checkTargetPerm(c *gin.Context, idents []string) {
	user := c.MustGet("user").(*models.User)
	nopri, err := user.NopriIdents(idents)
	ginx.Dangerous(err)

	if len(nopri) > 0 {
		ginx.Bomb(http.StatusForbidden, "No permission to operate the targets: %s", strings.Join(nopri, ", "))
	}
}
