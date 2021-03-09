package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type ModelTpl struct {
	PackageName    string
	ImportList     map[string]bool
	CreateTableDDL template.HTML
	DbName         string
	TableName      string
	ModelName      string
	RowsList       []template.HTML
}

func main() {

	//查询库名
	row := db.Raw("select database()").Row()
	var dbName string
	row.Scan(&dbName)
	//查询表名
	rows, err := db.Raw("SHOW TABLES").Rows()
	defer rows.Close()
	if err != nil {
		panic(err)
	}
	var tableList = make(map[string]bool)
	for rows.Next() {
		var tableName string
		rows.Scan(&tableName)
		//分表则去除后缀
		splitReg := regexp.MustCompile("_[0-9]+$")
		tableName = splitReg.ReplaceAllString(tableName, "")
		tableList[tableName] = true
	}
	//同步到模型
	dirName := "./model"
	_, err = os.Stat(dirName)
	if os.IsNotExist(err) {
		err = os.Mkdir(dirName, os.ModePerm)
		if err != nil {
			panic(err)
		}
	} else {
		//备份
		oldDirName := "./old_model"
		_, err = os.Stat(oldDirName)
		if os.IsNotExist(err) {
			err = os.Mkdir(oldDirName, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}
		err = os.RemoveAll(oldDirName)
		if err != nil {
			panic(err)
		}
		err = os.Rename(dirName, oldDirName)
		if err != nil {
			panic(err)
		}
		err = os.Mkdir(dirName, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	for k := range tableList {
		m := CreateModel(dbName, k)
		//存储到文件
		tpl, err := ioutil.ReadFile("./model.tpl")
		f, err := os.OpenFile(fmt.Sprintf("%v/%v_%v.go", dirName, dbName, k), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}
		t, err := template.New("model.tpl").Parse(string(tpl))
		if err != nil {
			panic(err)
		}
		err = t.Execute(f, m)
		cmd := exec.Command("gofmt", "-w", fmt.Sprintf("%v/%v_%v.go", dirName, dbName, k))
		cmd.Run()
		f.Close()
	}
}

func CreateModel(dbName, tableName string) ModelTpl {
	var modelTpl = ModelTpl{
		PackageName:    "model",
		ImportList:     map[string]bool{},
		CreateTableDDL: "",
		DbName:         dbName,
		TableName:      tableName,
		ModelName:      "",
		RowsList:       nil,
	}
	row := db.Raw(fmt.Sprintf("SHOW CREATE TABLE %v", tableName)).Row()
	var tmpName string
	var ddl string
	row.Scan(&tmpName, &ddl)
	modelTpl.CreateTableDDL = template.HTML(ddl)
	modelTpl.ModelName = FormatName(tableName)
	//解析ddl
	ds := strings.Split(ddl, "\n")
	for _, v := range ds {
		s := strings.TrimSpace(v)
		if len(s) > 0 {
			//第一个是否为字段名
			if s[0] == '`' {
				pos := strings.Index(s[1:], "`")
				modelFieldName := FormatName(s[1 : pos+1])
				mysqlFieldName := s[1 : pos+1]
				//第二个为字段类型
				s = s[pos+3:]
				pos = strings.Index(s, " ")
				dataTypeRegx := regexp.MustCompile(`^([a-zA-Z]*)`)
				t := dataTypeRegx.FindString(s[:pos])
				unsignedRegx := regexp.MustCompile(`(unsigned)`)
				isUnsigned := unsignedRegx.MatchString(v)
				notNullRegx := regexp.MustCompile(`(NOT NULL)`)
				isNotNull := notNullRegx.MatchString(v)
				l := map[string]string{
					"char":              "string",
					"varchar":           "string",
					"text":              "string",
					"mediumtext":        "string",
					"longtext":          "string",
					"bigint_unsigned":   "uint64",
					"bigint":            "int64",
					"int_unsigned":      "uint32",
					"int":               "int32",
					"smallint_unsigned": "uint16",
					"smallint":          "int16",
					"tinyint_unsigned":  "uint8",
					"tinyint":           "int8",
					"decimal":           "float64",
					"datetime":          "time.Time",
					"date":              "time.Time",
					"timestamp":         "time.Time",
				}
				var fieldType string
				if !isNotNull {
					fieldType = "*"
				}
				lk := t
				if isUnsigned {
					lk = fmt.Sprintf("%v_%v", t, "unsigned")
				}
				if t == "datetime" || t == "timestamp" || t == "date" {
					modelTpl.ImportList["time"] = true
				}
				fieldType += l[lk]
				modelTpl.RowsList = append(modelTpl.RowsList, template.HTML(fmt.Sprintf("%v %v `gorm:\"column:%v;type:%v\"`\n", modelFieldName, fieldType, mysqlFieldName, strings.Trim(s, ","))))
			}
		}
	}

	return modelTpl
}

//统一变量名称
func FormatName(tableName string) string {
	tableName = strings.ReplaceAll(tableName, "_", " ")
	tableName = strings.Title(tableName)
	tableName = strings.ReplaceAll(tableName, " ", "")
	return tableName
}
