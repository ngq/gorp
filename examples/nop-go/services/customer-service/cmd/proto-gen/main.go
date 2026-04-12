package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/provider/proto"
)

func main() {
	// 创建 Proto 生成器
	generator, err := proto.NewGenerator(&contract.ProtoGeneratorConfig{
		Enabled:              true,
		Strategy:             "protoc",
		DefaultProtoDir:      "./proto",
		IncludeHTTPAnnotation: true,
	})
	if err != nil {
		log.Fatalf("创建生成器失败: %v", err)
	}

	// 从 Service 接口生成 Proto
	err = generator.GenFromService(context.Background(), contract.ServiceToProtoOptions{
		ServicePath: "internal/service/customer_service.go",
		OutputPath:  "proto/customer.proto",
		Package:     "customer",
		GoPackage:   "nop-go/services/customer-service/proto;customer",
		ServiceName: "CustomerServiceRPC",
		IncludeHTTP: true,
		HTTPAnnotations: map[string]contract.HTTPRule{
			"Register":       {Method: "POST", Path: "/v1/customers", Body: "*"},
			"Login":          {Method: "POST", Path: "/v1/customers/login", Body: "*"},
			"GetCustomer":    {Method: "GET", Path: "/v1/customers/{id}"},
			"UpdateProfile":  {Method: "PUT", Path: "/v1/customers/{id}", Body: "*"},
			"ChangePassword": {Method: "POST", Path: "/v1/customers/{id}/password", Body: "*"},
			"ListCustomers":  {Method: "GET", Path: "/v1/customers"},
			"ValidateCustomer": {Method: "POST", Path: "/v1/customers/validate", Body: "*"},
		},
	})

	if err != nil {
		log.Fatalf("生成 Proto 失败: %v", err)
	}

	fmt.Println("✅ Proto 文件生成成功: ./proto/customer.proto")

	// 生成 AddressService Proto
	err = generator.GenFromService(context.Background(), contract.ServiceToProtoOptions{
		ServicePath: "internal/service/customer_service.go",
		OutputPath:  "proto/address.proto",
		Package:     "address",
		GoPackage:   "nop-go/services/customer-service/proto;address",
		ServiceName: "AddressServiceRPC",
		IncludeHTTP: true,
		HTTPAnnotations: map[string]contract.HTTPRule{
			"CreateAddress":      {Method: "POST", Path: "/v1/addresses", Body: "*"},
			"GetAddress":         {Method: "GET", Path: "/v1/addresses/{id}"},
			"ListAddresses":      {Method: "GET", Path: "/v1/customers/{customer_id}/addresses"},
			"UpdateAddress":      {Method: "PUT", Path: "/v1/addresses/{id}", Body: "*"},
			"DeleteAddress":      {Method: "DELETE", Path: "/v1/addresses/{id}"},
			"SetDefaultBilling":  {Method: "POST", Path: "/v1/addresses/{address_id}/default-billing", Body: "*"},
			"SetDefaultShipping": {Method: "POST", Path: "/v1/addresses/{address_id}/default-shipping", Body: "*"},
		},
	})

	if err != nil {
		log.Fatalf("生成 Address Proto 失败: %v", err)
	}

	fmt.Println("✅ Proto 文件生成成功: ./proto/address.proto")
}