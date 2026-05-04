package main

import (
	"context"
	"fmt"
	"log"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"github.com/ngq/gorp/framework/provider/proto"
)

func main() {
	// 鍒涘缓 Proto 鐢熸垚鍣?
	generator, err := proto.NewGenerator(&integrationcontract.ProtoGeneratorConfig{
		Enabled:               true,
		Strategy:              "protoc",
		DefaultProtoDir:       "./proto",
		IncludeHTTPAnnotation: true,
	})
	if err != nil {
		log.Fatalf("鍒涘缓鐢熸垚鍣ㄥけ璐? %v", err)
	}

	// 浠?Service 鎺ュ彛鐢熸垚 Proto
	err = generator.GenFromService(context.Background(), integrationcontract.ServiceToProtoOptions{
		ServicePath: "internal/service/customer_service.go",
		OutputPath:  "proto/customer.proto",
		Package:     "customer",
		GoPackage:   "nop-go/services/customer-service/proto;customer",
		ServiceName: "CustomerServiceRPC",
		IncludeHTTP: true,
		HTTPAnnotations: map[string]integrationcontract.HTTPRule{
			"Register":         {Method: "POST", Path: "/v1/customers", Body: "*"},
			"Login":            {Method: "POST", Path: "/v1/customers/login", Body: "*"},
			"GetCustomer":      {Method: "GET", Path: "/v1/customers/{id}"},
			"UpdateProfile":    {Method: "PUT", Path: "/v1/customers/{id}", Body: "*"},
			"ChangePassword":   {Method: "POST", Path: "/v1/customers/{id}/password", Body: "*"},
			"ListCustomers":    {Method: "GET", Path: "/v1/customers"},
			"ValidateCustomer": {Method: "POST", Path: "/v1/customers/validate", Body: "*"},
		},
	})

	if err != nil {
		log.Fatalf("鐢熸垚 Proto 澶辫触: %v", err)
	}

	fmt.Println("鉁?Proto 鏂囦欢鐢熸垚鎴愬姛: ./proto/customer.proto")

	// 鐢熸垚 AddressService Proto
	err = generator.GenFromService(context.Background(), integrationcontract.ServiceToProtoOptions{
		ServicePath: "internal/service/customer_service.go",
		OutputPath:  "proto/address.proto",
		Package:     "address",
		GoPackage:   "nop-go/services/customer-service/proto;address",
		ServiceName: "AddressServiceRPC",
		IncludeHTTP: true,
		HTTPAnnotations: map[string]integrationcontract.HTTPRule{
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
		log.Fatalf("鐢熸垚 Address Proto 澶辫触: %v", err)
	}

	fmt.Println("鉁?Proto 鏂囦欢鐢熸垚鎴愬姛: ./proto/address.proto")
}
