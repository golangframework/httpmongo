package httpmongo

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/golangframework/JSON"
	"github.com/golangframework/Object"
	"github.com/golangframework/moeregexp"
	"gopkg.in/mgo.v2/bson"
)

const (
	httpRequestBody = "&httprequestbody"
)

func dbo_Mongo_DB_C(w http.ResponseWriter, r *http.Request) error {
	var cmd = ""
	if moeregexp.IsMatch(Mongo_DB_C_func_path, r.URL.Path) {

		DB, C, cmd, _ := Mongo_DB_C_parse(r.URL.Path)

		funcname := cmd[0:strings.Index(cmd, "(")]
		args := cmd[strings.Index(cmd, "(")+1 : len(cmd)-1]

		if strings.Contains(args, httpRequestBody) {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic("没有 接收到 post数据。url不应该有" + httpRequestBody)
			}
			log.Print(string(body))
			args = strings.Replace(args, httpRequestBody, string(body), -1)
		}

		switch funcname {
		case "count":
			count(DB, C, w)
		case "find":
			find(DB, C, args, w)
		case "findcount":
			findcount(DB, C, args, w)
		case "findOne":
			findOne(DB, C, args, w)
		case "insert":
			insert(DB, C, args, w)
		case "insertmany":
			insertmany(DB, C, args, w)
		case "remove":
			remove(DB, C, args, w)
		case "save":
			save(DB, C, args, w)
		case "update":
			update(DB, C, args, w)
		default:
			return errors.New("请求函数名未知")
		}
		return errors.New("请求命令不支持")

	} else {
		w.Write([]byte("请求不匹配" + cmd))
		return errors.New("请求不匹配")
	}
}
func count(DB string, C string, w http.ResponseWriter) {
	c := MgoDataCollect(DB, C)
	count_, _ := c.Count()
	out := Object.Tostring(count_)
	w.Write([]byte(out))
}
func find(DB string, C string, args string, w http.ResponseWriter) {
	c := MgoDataCollect(DB, C)
	js := []JSON.JSON{} //结果集合

	var filter JSON.JSON
	if args == "" {
		args = "{}"
	}
	err := json.Unmarshal([]byte(args), &filter)
	if err != nil {
		panic("无法 序列化为 json")
	}

	err = c.Find(&filter).All(&js)

	jsonlist := JSON.ToJsonarraystring(js) //结果json字符串集合
	out := "[" + strings.Join(jsonlist, ",") + "]"
	w.Write([]byte(out))
}
func findcount(DB string, C string, args string, w http.ResponseWriter) {
	c := MgoDataCollect(DB, C)

	var filter JSON.JSON
	err := json.Unmarshal([]byte(args), &filter)
	if err != nil {
		panic("无法 序列化为 json")
	}

	count_, err := c.Find(&filter).Count()
	out := Object.Tostring(count_)
	w.Write([]byte(out))
}
func findOne(DB string, C string, args string, w http.ResponseWriter) {
	c := MgoDataCollect(DB, C)
	js := JSON.JSON{} //结果

	var filter JSON.JSON
	err := json.Unmarshal([]byte(args), &filter)
	if err != nil {
		panic("无法 序列化为 json")
	}

	err = c.Find(&filter).One(&js)

	out := js.ToJsonstring()
	w.Write([]byte(out))
}
func insert(DB string, C string, args string, w http.ResponseWriter) {
	c := MgoDataCollect(DB, C)
	var inserter JSON.JSON
	err := json.Unmarshal([]byte(args), &inserter)
	if err != nil {
		panic("无法序列化为 json")
	}
	err = c.Insert(inserter)
	if err != nil {
		panic("插入失败")
	}
	out := "{\"nInsert\":1}"
	w.Write([]byte(out))
}
func insertmany(DB string, C string, args string, w http.ResponseWriter) {
	c := MgoDataCollect(DB, C)
	/*args
	[

	{"name":"lipeng"},{"name":"test"}

	]
	*/
	var inserterlist []interface{}
	var ds = "[" + args + "]"
	log.Print(ds)
	err := json.Unmarshal([]byte(ds), &inserterlist)
	if err != nil {
		panic("无法 序列化为 []json")
	}

	err = c.Insert(inserterlist...)
	if err != nil {
		panic("批量插入失败")
	}

	out := "{\"nInsert\":" + Object.Tostring(len(inserterlist)) + "}"
	w.Write([]byte(out))
}
func remove(DB string, C string, args string, w http.ResponseWriter) {
	if args == "" {
		out := "{\"nRemoved\":0}"
		w.Write([]byte(out))
	} else {
		c := MgoDataCollect(DB, C)
		var filter JSON.JSON
		err := json.Unmarshal([]byte(args), &filter)
		if err != nil {
			panic("条件无法序列化为 json")
		}
		removeinfo, err := c.RemoveAll(filter)
		if err != nil {
			panic("删除失败")
		}
		out := "{\"nRemove\":" + Object.Tostring(removeinfo.Removed) + "}"
		w.Write([]byte(out))
	}
}
func save(DB string, C string, args string, w http.ResponseWriter) {
	if args == "" {
		out := "{\"nInsert\":0,\"nUpdate\":0}"
		w.Write([]byte(out))
	} else {
		c := MgoDataCollect(DB, C)
		var saver bson.M
		err := json.Unmarshal([]byte(args), &saver)
		out := ""
		if err != nil {
			panic("条件无法序列化为 json")
		}

		if saver["_id"] == nil {
			err = c.Insert(saver)
			if err != nil {
				panic("插入失败")
			}
			out = "{\"nInsert\":1}"
		} else {
			if bson.IsObjectIdHex(Object.Tostring(saver["_id"])) == false {
				panic("saver 中 _id 不正确")
			} else {
				filter := bson.M{"_id": bson.ObjectIdHex(Object.Tostring(saver["_id"]))}

				count, _ := c.Find(filter).Count()
				saver["_id"] = bson.ObjectIdHex(Object.Tostring(saver["_id"]))

				if count >= 1 {
					rinfo, err := c.RemoveAll(filter)
					if err != nil {
						panic("删除失败" + JSON.ToJsonstring(rinfo))
					}
					err = c.Insert(saver)
					if err != nil {
						panic("插入失败")
					}
					out = "{\"nUpdate\":1}"
				} else {
					err = c.Insert(saver)
					if err != nil {
						panic("插入失败")
					}
					out = "{\"nInsert\":1}"
				}
			}
		}

		w.Write([]byte(out))
	}
}
func update(DB string, C string, args string, w http.ResponseWriter) {
	if args == "" {
		out := "{\"nUpdate\":0}"
		w.Write([]byte(out))
	} else {
		/*args
		[

		{"name":"lipeng"},{"name":"test"}

		]
		*/
		var updatearg []JSON.JSON
		err := json.Unmarshal([]byte("["+args+"]"), &updatearg)
		if err != nil || len(updatearg) != 2 {
			panic("条件无法序列化为 2个json")
		}
		c := MgoDataCollect(DB, C)
		log.Print(updatearg[0], updatearg[1])
		updateinfo, err := c.UpdateAll(updatearg[0], updatearg[1])
		if err != nil {
			panic("更新失败")
		}
		out := "{\"nUpdate\":" + Object.Tostring(updateinfo.Updated) + "}"
		w.Write([]byte(out))
	}
}
