## How generates Go code from a WSDL file

`go install github.com/hooklift/gowsdl/cmd/gowsdl@latest`
```
Usage: gowsdl [options] myservice.wsdl
  -o string
        File where the generated code will be saved (default "myservice.go")
  -p string
        Package under which code will be generated (default "myservice")
  -i    Skips TLS Verification
  -v    Shows gowsdl version
```
**注意:** 对于onvif的wsdl文件需要注释引入的文件，防止引入的文件中有与当前文件重名类型。基础类型在xsd文件夹中。