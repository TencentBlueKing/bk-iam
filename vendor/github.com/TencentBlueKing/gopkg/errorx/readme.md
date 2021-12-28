# errorx

`errorx`实现了一个自定义error, 可以进行错误详情wrap.

封装的`Wrap/Wrapf`存在两个参数: `layer`层级, `function`函数名, 可以方便地定位报错调用链的层级及函数.

最终效果:

```
system error[request_id=0838c4090cfa4f4f9dceea5fd8c7b029]: [Handler:validateSystemSuperUser] impls.ListSubjectRoleSystemID subjectType=`user`, subjectID=`test1` fail%!(EXTRA string=test1) => [Cache:GetSubjectRole] SubjectRoleCache.Get subjectType=`user`, subjectID=`test1` fail => [Cache:GetSubjectPK] SubjectPKCache.Get _type=`user`, id=`test1` fail => [SubjectSVC:GetPK] GetPK _type=`user`, id=`test1` fail => [Raw:Error] sql: no rows in result set
```


## Usage

### 1. 直接wrap


```go
import "github.com/TencentBlueKing/gopkg/errorx"

cnt, err := l.relationManager.GetMemberCount(_type, id)
if err != nil {
    return errorx.Wrapf(err, "ServiceLayer", "GetMemberCount",
            "relationManager.GetMemberCount _type=`%s`, id=`%s` fail", _type, id)
}
```

### 2. 函数中存在多个return


```go
import "github.com/TencentBlueKing/gopkg/errorx"

// create a func with layer name and function name
errorWrapf := errorx.NewLayerFunctionErrorWrapf("ServiceLayer", "BulkDeleteSubjectMember")

if err != nil {
    return errorWrapf(err, "relationManager.UpdateExpiredAt relations=`%+v` fail", relations)
}

// ...

if err != nil {
    return errorWrapf(err, "relationManager.DoSomething relations=`%+v` fail", relations)
}
```
