package models

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/ccfos/nightingale/v6/pkg/ldapx"
	"github.com/ccfos/nightingale/v6/pkg/ormx"
	"github.com/ccfos/nightingale/v6/pkg/poster"

	"errors"

	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/slice"
	"github.com/toolkits/pkg/str"
)

const (
	Dingtalk     = "dingtalk"
	Wecom        = "wecom"
	Feishu       = "feishu"
	FeishuCard   = "feishucard"
	Mm           = "mm"
	Telegram     = "telegram"
	Email        = "email"
	EmailSubject = "mailsubject"

	DingtalkKey = "dingtalk_robot_token"
	WecomKey    = "wecom_robot_token"
	FeishuKey   = "feishu_robot_token"
	MmKey       = "mm_webhook_url"
	TelegramKey = "telegram_robot_token"
)

var (
	DefaultChannels = []string{Dingtalk, Wecom, Feishu, Mm, Telegram, Email, FeishuCard}
)

const UserTableName = "users"

type User struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id       int64    `json:"id" column:"id"`
	Username string   `json:"username" column:"username"`
	Nickname string   `json:"nickname" column:"nickname"`
	Password string   `json:"-" column:"password"`
	Phone    string   `json:"phone" column:"phone"`
	Email    string   `json:"email" column:"email"`
	Portrait string   `json:"portrait" column:"portrait"`
	Roles    string   `json:"-" column:"roles"` // 这个字段写入数据库
	RolesLst []string `json:"roles"`            // 这个字段和前端交互
	// Contacts   ormx.JSONObj `json:"contacts" column:"contacts"`     // 内容为 map[string]string 结构
	Contacts     string       `json:"-" column:"contacts"`            // 内容为 map[string]string 结构
	ContactsJson ormx.JSONObj `json:"contacts"`                       // 内容为 map[string]string 结构
	Maintainer   int          `json:"maintainer" column:"maintainer"` // 是否给管理员发消息 0:not send 1:send
	CreateAt     int64        `json:"create_at" column:"create_at"`
	CreateBy     string       `json:"create_by" column:"create_by"`
	UpdateAt     int64        `json:"update_at" column:"update_at"`
	UpdateBy     string       `json:"update_by" column:"update_by"`
	Admin        bool         `json:"admin"` // 方便前端使用
}

func (u *User) GetTableName() string {
	return UserTableName
}

func (u *User) DB2FE() error {
	return nil
}

func (u *User) String() string {
	// bs, err := u.Contacts.MarshalJSON()
	// if err != nil {
	// 	return err.Error()
	// }
	bs := u.Contacts

	return fmt.Sprintf("<id:%d username:%s nickname:%s email:%s phone:%s contacts:%s>", u.Id, u.Username, u.Nickname, u.Email, u.Phone, string(bs))
}

func (u *User) IsAdmin() bool {
	for i := 0; i < len(u.RolesLst); i++ {
		if u.RolesLst[i] == AdminRole {
			return true
		}
	}
	return false
}

func (u *User) Verify() error {
	u.Username = strings.TrimSpace(u.Username)

	if u.Username == "" {
		return errors.New("Username is blank")
	}

	if str.Dangerous(u.Username) {
		return errors.New("Username has invalid characters")
	}

	if str.Dangerous(u.Nickname) {
		return errors.New("Nickname has invalid characters")
	}

	if u.Phone != "" && !str.IsPhone(u.Phone) {
		return errors.New("Phone invalid")
	}

	if u.Email != "" && !str.IsMail(u.Email) {
		return errors.New("Email invalid")
	}

	return nil
}

func (u *User) Add(ctx *ctx.Context) error {
	user, err := UserGetByUsername(ctx, u.Username)
	if err != nil {
		return fmt.Errorf("failed to query user:%w", err)
	}

	if user != nil {
		return errors.New("Username already exists")
	}

	now := time.Now().Unix()
	u.CreateAt = now
	u.UpdateAt = now
	return Insert(ctx, u)
}

func (u *User) Update(ctx *ctx.Context, selectFields ...string) error {
	if err := u.Verify(); err != nil {
		return err
	}
	return Update(ctx, u, selectFields)
	//return DB(ctx).Model(u).Select(selectField, selectFields...).Updates(u).Error
}

func (u *User) UpdateAllFields(ctx *ctx.Context) error {
	if err := u.Verify(); err != nil {
		return err
	}

	u.UpdateAt = time.Now().Unix()
	return Update(ctx, u, nil)
	//return DB(ctx).Model(u).Select("*").Updates(u).Error
}

func (u *User) UpdatePassword(ctx *ctx.Context, password, updateBy string) error {
	finder := zorm.NewUpdateFinder(UserTableName).Append("password=?,update_at=?,update_by=? WHERE id=?", password, time.Now().Unix(), updateBy, u.Id)
	return UpdateFinder(ctx, finder)
	/*
		return DB(ctx).Model(u).Updates(map[string]interface{}{
			"password":  password,
			"update_at": time.Now().Unix(),
			"update_by": updateBy,
		}).Error
	*/
}

func (u *User) Del(ctx *ctx.Context) error {
	_, err := zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		f1 := zorm.NewDeleteFinder(UserGroupMemberTableName).Append("WHERE user_id=?", u.Id)
		_, err := zorm.UpdateFinder(ctx, f1)
		if err != nil {
			return nil, err
		}
		return zorm.Delete(ctx, u)
	})
	return err
	/*
		return DB(ctx).Transaction(func(tx *zorm.DBDao) error {
			if err := tx.Where("user_id=?", u.Id).Delete(&UserGroupMember{}).Error; err != nil {
				return err
			}

			if err := tx.Where("id=?", u.Id).Delete(&User{}).Error; err != nil {
				return err
			}

			return nil
		})
	*/
}

func (u *User) ChangePassword(ctx *ctx.Context, oldpass, newpass string) error {
	_oldpass, err := CryptoPass(ctx, oldpass)
	if err != nil {
		return err
	}

	_newpass, err := CryptoPass(ctx, newpass)
	if err != nil {
		return err
	}

	if u.Password != _oldpass {
		return errors.New("Incorrect old password")
	}

	return u.UpdatePassword(ctx, _newpass, u.Username)
}

func UserGet(ctx *ctx.Context, where string, args ...interface{}) (*User, error) {
	lst := make([]*User, 0)
	finder := zorm.NewSelectFinder(UserTableName)
	AppendWhere(finder, where, args...)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where(where, args...).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}

	lst[0].RolesLst = strings.Fields(lst[0].Roles)
	lst[0].Admin = lst[0].IsAdmin()
	lst[0].ContactsJson.Scan(lst[0].Contacts)

	return lst[0], nil
}

func UserGetByUsername(ctx *ctx.Context, username string) (*User, error) {
	return UserGet(ctx, "username=?", username)
}

func UserGetById(ctx *ctx.Context, id int64) (*User, error) {
	return UserGet(ctx, "id=?", id)
}

func InitRoot(ctx *ctx.Context) {
	user, err := UserGetByUsername(ctx, "root")
	if err != nil {
		fmt.Println("failed to query user root:", err)
		os.Exit(1)
	}

	if user == nil {
		return
	}

	if len(user.Password) > 31 {
		// already done before
		return
	}

	newPass, err := CryptoPass(ctx, user.Password)
	if err != nil {
		fmt.Println("failed to crypto pass:", err)
		os.Exit(1)
	}
	err = UpdateColumn(ctx, UserTableName, user.Id, "password", newPass)
	//err = DB(ctx).Model(user).Update("password", newPass).Error
	if err != nil {
		fmt.Println("failed to update root password:", err)
		os.Exit(1)
	}

	fmt.Println("root password init done")
}

func PassLogin(ctx *ctx.Context, username, pass string) (*User, error) {
	user, err := UserGetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("Username or password invalid")
	}

	loginPass, err := CryptoPass(ctx, pass)
	if err != nil {
		return nil, err
	}
	logger.Infof("loginPass: %s", loginPass)
	logger.Infof("user.Password: %s", user.Password)

	if loginPass != user.Password {
		return nil, fmt.Errorf("Username or password invalid")
	}

	return user, nil
}

func LdapLogin(ctx *ctx.Context, username, pass, roles string, ldap *ldapx.SsoClient) (*User, error) {
	sr, err := ldap.LdapReq(username, pass)
	if err != nil {
		return nil, err
	}

	user, err := UserGetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		// default user settings
		user = &User{
			Username: username,
			Nickname: username,
		}
	}

	// copy attributes from ldap
	ldap.RLock()
	attrs := ldap.Attributes
	coverAttributes := ldap.CoverAttributes
	ldap.RUnlock()

	if attrs.Nickname != "" {
		user.Nickname = sr.Entries[0].GetAttributeValue(attrs.Nickname)
	}
	if attrs.Email != "" {
		user.Email = sr.Entries[0].GetAttributeValue(attrs.Email)
	}
	if attrs.Phone != "" {
		user.Phone = strings.Replace(sr.Entries[0].GetAttributeValue(attrs.Phone), " ", "", -1)
	}

	if user.Roles == "" {
		user.Roles = roles
	}

	if user.Id > 0 {
		if coverAttributes {
			_, err := zorm.Update(ctx.Ctx, user)
			//err := DB(ctx).Updates(user).Error
			if err != nil {
				return nil, fmt.Errorf("failed to update user:%w", err)
			}
		}
		return user, nil
	}

	now := time.Now().Unix()

	user.Password = "******"
	user.Portrait = ""

	// user.Contacts = []byte("{}")
	user.Contacts = "{}"
	user.CreateAt = now
	user.UpdateAt = now
	user.CreateBy = "ldap"
	user.UpdateBy = "ldap"
	_, err = zorm.Insert(ctx.Ctx, user)
	//err = DB(ctx).Create(user).Error
	return user, err
}

func UserTotal(ctx *ctx.Context, query string) (num int64, err error) {
	finder := zorm.NewSelectFinder(UserTableName, "count(*)")
	if query != "" {
		q := "%" + query + "%"
		finder.Append("WHERE username like ? or nickname like ? or phone like ? or email like ?", q, q, q, q)
		//num, err = Count(DB(ctx).Model(&User{}).Where("username like ? or nickname like ? or phone like ? or email like ?", q, q, q, q))
	} //else {
	//	num, err = Count(DB(ctx).Model(&User{}))
	//}
	num, err = Count(ctx, finder)
	if err != nil {
		return num, fmt.Errorf("failed to count user:%w", err)
	}

	return num, nil
}

func UserGets(ctx *ctx.Context, query string, limit, offset int) ([]User, error) {
	finder := zorm.NewSelectFinder(UserTableName)
	finder.SelectTotalCount = false
	page := zorm.NewPage()
	page.PageSize = limit
	page.PageNo = offset / limit
	//session := DB(ctx).Limit(limit).Offset(offset).Order("username")
	if query != "" {
		q := "%" + query + "%"
		finder.Append("WhERE username like ? or nickname like ? or phone like ? or email like ?", q, q, q, q)
		//session = session.Where("username like ? or nickname like ? or phone like ? or email like ?", q, q, q, q)
	}
	finder.Append("order by username asc ")

	users := make([]User, 0)
	err := zorm.Query(ctx.Ctx, finder, &users, page)
	//err := session.Find(&users).Error
	if err != nil {
		return users, fmt.Errorf("failed to query user:%w", err)
	}

	for i := 0; i < len(users); i++ {
		users[i].RolesLst = strings.Fields(users[i].Roles)
		users[i].Admin = users[i].IsAdmin()
		users[i].Password = ""
		users[i].ContactsJson.Scan(users[i].Contacts)
	}

	return users, nil
}

func UserGetAll(ctx *ctx.Context) ([]*User, error) {
	if !ctx.IsCenter {
		lst, err := poster.GetByUrls[[]*User](ctx, "/v1/n9e/users")
		return lst, err
	}

	lst := make([]*User, 0)
	finder := zorm.NewSelectFinder(UserTableName)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Find(&lst).Error
	if err == nil {
		for i := 0; i < len(lst); i++ {
			lst[i].RolesLst = strings.Fields(lst[i].Roles)
			lst[i].Admin = lst[i].IsAdmin()
		}
	}
	return lst, err
}

func UserGetsByIds(ctx *ctx.Context, ids []int64) ([]User, error) {
	if len(ids) == 0 {
		return []User{}, nil
	}

	lst := make([]User, 0)
	finder := zorm.NewSelectFinder(UserTableName).Append("WHERE id in (?) order by username asc", ids)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where("id in ?", ids).Order("username").Find(&lst).Error
	if err == nil {
		for i := 0; i < len(lst); i++ {
			lst[i].RolesLst = strings.Fields(lst[i].Roles)
			lst[i].Admin = lst[i].IsAdmin()
		}
	}

	return lst, err
}

func (u *User) CanModifyUserGroup(ctx *ctx.Context, ug *UserGroup) (bool, error) {
	// 我是管理员，自然可以
	if u.IsAdmin() {
		return true, nil
	}

	// 我是创建者，自然可以
	if ug.CreateBy == u.Username {
		return true, nil
	}

	// 我是成员，也可以吧，简单搞
	num, err := UserGroupMemberCount(ctx, "user_id=? and group_id=?", u.Id, ug.Id)
	if err != nil {
		return false, err
	}

	return num > 0, nil
}

func (u *User) CanDoBusiGroup(ctx *ctx.Context, bg *BusiGroup, permFlag ...string) (bool, error) {
	if u.IsAdmin() {
		return true, nil
	}

	// 我在任意一个UserGroup里，就有权限
	ugids, err := UserGroupIdsOfBusiGroup(ctx, bg.Id, permFlag...)
	if err != nil {
		return false, err
	}

	if len(ugids) == 0 {
		return false, nil
	}

	num, err := UserGroupMemberCount(ctx, "user_id = ? and group_id in (?)", u.Id, ugids)
	return num > 0, err
}

func (u *User) CheckPerm(ctx *ctx.Context, operation string) (bool, error) {
	if u.IsAdmin() {
		return true, nil
	}

	return RoleHasOperation(ctx, u.RolesLst, operation)
}

func UserStatistics(ctx *ctx.Context) (*Statistics, error) {
	if !ctx.IsCenter {
		s, err := poster.GetByUrls[*Statistics](ctx, "/v1/n9e/statistic?name=user")
		return s, err
	}
	return StatisticsGet(ctx, UserTableName)
	/*
		session := DB(ctx).Model(&User{}).Select("count(*) as total", "max(update_at) as last_updated")

		var stats []*Statistics
		err := session.Find(&stats).Error
		if err != nil {
			return nil, err
		}

		return stats[0], nil
	*/
}

func (u *User) NopriIdents(ctx *ctx.Context, idents []string) ([]string, error) {
	if u.IsAdmin() {
		return []string{}, nil
	}

	ugids, err := MyGroupIds(ctx, u.Id)
	if err != nil {
		return []string{}, err
	}

	if len(ugids) == 0 {
		return idents, nil
	}

	bgids, err := BusiGroupIds(ctx, ugids, "rw")
	if err != nil {
		return []string{}, err
	}

	if len(bgids) == 0 {
		return idents, nil
	}

	arr := make([]string, 0)
	finder := zorm.NewSelectFinder(TargetTableName, "ident").Append("WHERE group_id in (?)", bgids)
	err = zorm.Query(ctx.Ctx, finder, &arr, nil)
	//err = DB(ctx).Model(&Target{}).Where("group_id in ?", bgids).Pluck("ident", &arr).Error
	if err != nil {
		return []string{}, err
	}

	return slice.SubString(idents, arr), nil
}

// 我是管理员，返回所有
// 或者我是成员
func (u *User) BusiGroups(ctx *ctx.Context, limit int, query string, all ...bool) ([]BusiGroup, error) {
	finder := zorm.NewSelectFinder(BusiGroupTableName).Append("WHERE 1=1 ")
	finder.SelectTotalCount = false
	page := zorm.NewPage()
	page.PageSize = limit
	//session := DB(ctx).Order("name").Limit(limit)

	lst := make([]BusiGroup, 0)
	if u.IsAdmin() || (len(all) > 0 && all[0]) {
		finder.Append("and name like ? order by name asc", "%"+query+"%")
		err := zorm.Query(ctx.Ctx, finder, &lst, page)
		//err := session.Where("name like ?", "%"+query+"%").Find(&lst).Error
		if err != nil {
			return lst, err
		}

		if len(lst) == 0 && len(query) > 0 {
			// 隐藏功能，一般人不告诉，哈哈。query可能是给的ident，所以上面的sql没有查到，当做ident来查一下试试
			var t *Target
			t, err = TargetGet(ctx, "ident=?", query)
			if err != nil {
				return lst, err
			}

			if t == nil {
				return lst, nil
			}
			finder := zorm.NewSelectFinder(BusiGroupTableName).Append("WHERE id=? order by name asc", t.GroupId)
			finder.SelectTotalCount = false
			err = zorm.Query(ctx.Ctx, finder, &lst, page)
			//err = DB(ctx).Order("name").Limit(limit).Where("id=?", t.GroupId).Find(&lst).Error
		}

		return lst, err
	}

	userGroupIds, err := MyGroupIds(ctx, u.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get MyGroupIds:%w", err)
	}

	busiGroupIds, err := BusiGroupIds(ctx, userGroupIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get BusiGroupIds:%w", err)
	}

	if len(busiGroupIds) == 0 {
		return lst, nil
	}
	finder.Append("and id in (?) and name like ? order by name asc", busiGroupIds, "%"+query+"%")
	err = zorm.Query(ctx.Ctx, finder, &lst, page)
	//err = session.Where("id in ?", busiGroupIds).Where("name like ?", "%"+query+"%").Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 && len(query) > 0 {
		var t *Target
		t, err = TargetGet(ctx, "ident=?", query)
		if err != nil {
			return lst, err
		}

		if slice.ContainsInt64(busiGroupIds, t.GroupId) {
			finder := zorm.NewSelectFinder(BusiGroupTableName).Append("WHERE id=? order by name asc", t.GroupId)
			finder.SelectTotalCount = false
			err = zorm.Query(ctx.Ctx, finder, &lst, page)
			//err = DB(ctx).Order("name").Limit(limit).Where("id=?", t.GroupId).Find(&lst).Error
		}
	}

	return lst, err
}

func (u *User) UserGroups(ctx *ctx.Context, limit int, query string) ([]UserGroup, error) {
	finder := zorm.NewSelectFinder(UserGroupTableName).Append("WHERE 1=1 ")
	finder.SelectTotalCount = false
	page := zorm.NewPage()
	page.PageSize = limit
	//session := DB(ctx).Order("name").Limit(limit)

	lst := make([]UserGroup, 0)
	if u.IsAdmin() {
		finder.Append("and name like ? order by name asc", "%"+query+"%")
		err := zorm.Query(ctx.Ctx, finder, &lst, page)
		//err := session.Where("name like ?", "%"+query+"%").Find(&lst).Error
		if err != nil {
			return lst, err
		}

		var user *User
		if len(lst) == 0 && len(query) > 0 {
			// 隐藏功能，一般人不告诉，哈哈。query可能是给的用户名，所以上面的sql没有查到，当做user来查一下试试
			user, err = UserGetByUsername(ctx, query)
			if user == nil {
				return lst, err
			}
			var ids []int64
			ids, err = MyGroupIds(ctx, user.Id)
			if err != nil || len(ids) == 0 {
				return lst, err
			}
			lst, err = UserGroupGetByIds(ctx, ids)
		}
		return lst, err
	}

	ids, err := MyGroupIds(ctx, u.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get MyGroupIds:%w", err)
	}

	if len(ids) > 0 {
		//session = session.Where("id in ? or create_by = ?", ids, u.Username)
		finder.Append("and (id in (?) or create_by = ?)", ids, u.Username)
	} else {
		//session = session.Where("create_by = ?", u.Username)
		finder.Append("and create_by = ?", u.Username)
	}

	if len(query) > 0 {
		//session = session.Where("name like ?", "%"+query+"%")
		finder.Append("and name like ?", "%"+query+"%")
	}
	finder.Append("order by name asc")
	err = zorm.Query(ctx.Ctx, finder, &lst, page)
	//err = session.Find(&lst).Error
	return lst, err
}

func (u *User) ExtractToken(key string) (string, bool) {
	// bs, err := u.Contacts.MarshalJSON()
	// if err != nil {
	// 	logger.Errorf("ExtractToken: failed to marshal contacts: %v", err)
	// 	return "", false
	// }
	var err error
	bs := []byte(u.Contacts)

	jsonMap := make(map[string]string, 0)
	err = json.Unmarshal(bs, &jsonMap)
	if err != nil {
		logger.Errorf("ExtractToken: failed to unmarshal contacts: %v", err)
		return "", false
	}
	value := ""
	ok := false
	switch key {
	case Dingtalk:
		//ret := gjson.GetBytes(bs, DingtalkKey)
		//return ret.String(), ret.Exists()
		value, ok = jsonMap[DingtalkKey]
	case Wecom:
		//ret := gjson.GetBytes(bs, WecomKey)
		//return ret.String(), ret.Exists()
		value, ok = jsonMap[WecomKey]
	case Feishu, FeishuCard:
		//ret := gjson.GetBytes(bs, FeishuKey)
		//return ret.String(), ret.Exists()
		value, ok = jsonMap[FeishuKey]
	case Mm:
		//ret := gjson.GetBytes(bs, MmKey)
		//return ret.String(), ret.Exists()
		value, ok = jsonMap[MmKey]
	case Telegram:
		//ret := gjson.GetBytes(bs, TelegramKey)
		//return ret.String(), ret.Exists()
		value, ok = jsonMap[TelegramKey]
	case Email:
		value = u.Email
		ok = u.Email != ""
		//return u.Email, u.Email != ""
	default:
		return "", false
	}
	return value, ok
}
