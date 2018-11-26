package main

import (
	"fmt"
  "regexp"

	gp "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/tcncloud/protoc-gen-state/state"
	"strings"
)

type SideEffect int
type Crud int
type CludgeEffect int

const (
	REQUEST         SideEffect = 0
	SUCCESS         SideEffect = 1
	FAILURE         SideEffect = 2
	CANCEL          SideEffect = 3
	SIDE_EFFECT_MAX SideEffect = 4
)

const (
	CREATE   Crud = 0
	GET      Crud = 1
	UPDATE   Crud = 2
	DELETE   Crud = 3
	CRUD_MAX Crud = 4
	CUSTOM   Crud = 5
)

type ModifiedFieldDescriptorProto struct {
  field *gp.FieldDescriptorProto
  message *gp.DescriptorProto
  file *gp.FileDescriptorProto
}

func CreatePackageAndTypeString(in string) string {
  // remove the first character if it's a period
  if in[0] == '.' {
    in = in[1:]
  }

  period := regexp.MustCompile("\\.")
  numPeriods := len(period.FindAllStringIndex(in, -1))

  // if there is only one period, the package name has no periods in it
  if numPeriods <= 1 {
    return in
  }

  // replace all but the last period with underscore
  return strings.Replace(in, ".", "_", numPeriods -1)
}

func SideEffectName(s SideEffect) string {
	switch s {
	case REQUEST:
		return "request"
	case SUCCESS:
		return "success"
	case FAILURE:
		return "failure"
	case CANCEL:
		return "cancel"
	default:
		return ""
	}
}

func CrudName(crud Crud, repeated bool) string {
	switch crud {
	case CREATE:
		return "create"
	case GET:
		{
			if repeated {
				return "list"
			} else {
				return "get"
			}
		}
	case UPDATE:
		return "update"
	case DELETE:
		return "delete"
	case CUSTOM:
		return "custom"
	default:
		return ""
	}
}

func Tabs(n int) string {
  output := ""
  for i:=0; i<n; i++ {
    output += "  "
  }
  return output
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func containsFile(s []*gp.FileDescriptorProto, f *gp.FileDescriptorProto) bool {
	for _, a := range s {
		if a.GetName() == f.GetName() {
			return true
		}
	}
	return false
}

func GetFilePath(name string) string {
	// replace proto filetype
	name = strings.Replace(name, ".proto", "_pb", 1)
	return name
}

func CreateAggregateByPackage(msgFiles []*gp.FileDescriptorProto, protocTsPath string, statePkg string) []*File {
	var packageNames []string
	var result string
	out := make([]*File, 0)

	for _, file := range msgFiles {
		if !contains(packageNames, file.GetPackage()) && statePkg != file.GetPackage() {
			result = ""
			packageNames = append(packageNames, file.GetPackage())

			for _, f := range msgFiles {
				if f.GetPackage() == file.GetPackage() {
					result += fmt.Sprintf("export * from \"%s%s\";\n", protocTsPath, GetFilePath(f.GetName()))
				}
			}

			out = append(out, &File{
				// replace package name periods with underscores
				Name:    fmt.Sprintf("%s_aggregate.ts", strings.Replace(file.GetPackage(), ".", "_", -1)),
				Content: result,
			})
		}
	}
	return out
}

// find a method descriptor from the annotation string name
func FindMethodDescriptor(serviceFiles []*gp.FileDescriptorProto, fullMethodName string) (*gp.MethodDescriptorProto, error) {
	for _, servFile := range serviceFiles {
		packageName := servFile.GetPackage()
		for _, service := range servFile.GetService() {
			serviceName := service.GetName()
			for _, method := range service.GetMethod() {
				if fmt.Sprintf("%s.%s.%s", packageName, serviceName, method.GetName()) == fullMethodName {
					// make sure it doesn't use client-side streaming (not supported with grpc-web)
					if method.GetClientStreaming() {
						return nil, fmt.Errorf("Client-side streaming not supported. Failed on method: %s", fullMethodName)
					}
					return method, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("Unable to locate method: \"%s\". Missing Method Declaration in Service.", fullMethodName)
}

func GetAnnotation(meth state.StringFieldOptions, crud Crud, repeated bool) string {
	switch crud {
	case CREATE:
		return meth.GetCreate()
	case GET:
		{
			if repeated {
				return meth.GetList()
			} else {
				return meth.GetGet()
			}
		}
	case UPDATE:
		return meth.GetUpdate()
	case DELETE:
		return meth.GetDelete()
	default:
		return ""
	}
}