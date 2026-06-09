//go:build integration

package brobridgetest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

// The senario of this Unit-test combo
// *daginwu*

// Table mean
//{Index number}.{Testcase name}
//{describe what happen in these testcase}

// 1. Create router pass
//    - create router: fred01
//    - create router: fred02
//    - create router: fred03
//    - create router: dagin01
//    - create router: dagin02
//    - create router: dagin03
//    - Print all UserID for checking

// 2. Create router fail
//    - create router fail: no name
//    - create router fail: no protocol
//    - create router fail: no namespace (TODO this api calling will refresh the id need to fix that)
//    - create router fail: create same router 2 times (TODO: 2 same router can't create again)

// 3. Get routers pass
//    - get router in one page
//    - get 2 router in per-page and flip page
//    - get all router by using "-1" keyword

// 4. Get routers fail
//    - no negative limit query

// 5. Enable router pass
//    - enable router
//    - disable router

// 6. Enable router fail
// 	  - no router id fail
// 	  - no enabled JSON fail (TODO: No request JSON need to be fail and response can be more pretty)

// 7. Delete router pass
//    - delete router

// 8. Delete router fail
//    - no router id fail

// 9. Update router pass
//    - update router

// 10. Update router fail
//	   - no router id fail
//	   - no update information fail

// 12. Get router mapping pass
//	   - get all gateway that map to router (TODO: fail api calling)

// 13. Get router mapping fail
//     - no router id fail

// 14. Update router mapping pass
//	   - mapping router

// 15. Update router mapping fail
// 	   - no gateway to map fail (TODO)

// Global variables that combo-test need
var testurl string = "http://127.0.0.1:22112"
var routers []string

// Create router testcase
func TestCreateRouterPass1(t *testing.T) {
	e := httpexpect.New(t, testurl)
	contentType := "application/json"
	postdata := map[string]interface{}{
		"name":        "fred01",
		"description": "hellofred",
		"protocol":    "http",
		"namespace":   "brobridge",
	}
	obj := e.POST("/api/v1/routers").
		WithHeader("Content-Type", contentType).
		WithJSON(postdata).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("id")
	routers = append(routers, obj.Value("id").String().Raw())
}

func TestCreateRouterPass2(t *testing.T) {
	e := httpexpect.New(t, testurl)
	contentType := "application/json"
	postdata := map[string]interface{}{
		"name":        "fred02",
		"description": "hellofred",
		"protocol":    "http",
		"namespace":   "brobridge",
	}
	obj := e.POST("/api/v1/routers").
		WithHeader("Content-Type", contentType).
		WithJSON(postdata).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("id")
	routers = append(routers, obj.Value("id").String().Raw())
}
func TestCreateRouterPass3(t *testing.T) {
	e := httpexpect.New(t, testurl)
	contentType := "application/json"
	postdata := map[string]interface{}{
		"name":        "fred02",
		"description": "hellofred",
		"protocol":    "http",
		"namespace":   "brobridge",
	}
	obj := e.POST("/api/v1/routers").
		WithHeader("Content-Type", contentType).
		WithJSON(postdata).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("id")
	routers = append(routers, obj.Value("id").String().Raw())
}

func TestCreateRouterPass4(t *testing.T) {
	e := httpexpect.New(t, testurl)
	contentType := "application/json"
	postdata := map[string]interface{}{
		"name":        "dagin01",
		"description": "hellofred",
		"protocol":    "http",
		"namespace":   "brobridge",
	}
	obj := e.POST("/api/v1/routers").
		WithHeader("Content-Type", contentType).
		WithJSON(postdata).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("id")
	routers = append(routers, obj.Value("id").String().Raw())
}

func TestCreateRouterPass5(t *testing.T) {
	e := httpexpect.New(t, testurl)
	contentType := "application/json"
	postdata := map[string]interface{}{
		"name":        "dagin02",
		"description": "hellofred",
		"protocol":    "http",
		"namespace":   "brobridge",
	}
	obj := e.POST("/api/v1/routers").
		WithHeader("Content-Type", contentType).
		WithJSON(postdata).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("id")
	routers = append(routers, obj.Value("id").String().Raw())
}

func TestCreateRouterPass6(t *testing.T) {
	e := httpexpect.New(t, testurl)
	contentType := "application/json"
	postdata := map[string]interface{}{
		"name":        "dagin03",
		"description": "hellofred",
		"protocol":    "http",
		"namespace":   "brobridge",
	}
	obj := e.POST("/api/v1/routers").
		WithHeader("Content-Type", contentType).
		WithJSON(postdata).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("id")
	routers = append(routers, obj.Value("id").String().Raw())

	// List all user
	fmt.Println("====================================")
	for _, user := range routers {

		fmt.Println(user)
	}
	fmt.Println("====================================")
}

func TestCreateRouterFail1(t *testing.T) {
	e := httpexpect.New(t, testurl)
	contentType := "application/json"
	postdata := map[string]interface{}{
		"name":        "",
		"description": "hellofred",
		"protocol":    "http",
		"namespace":   "brobridge",
	}
	e.POST("/api/v1/routers").
		WithHeader("Content-Type", contentType).
		WithJSON(postdata).
		Expect().
		Status(http.StatusConflict).
		JSON().
		Object().
		ContainsKey("error")
}

func TestCreateRouterFail2(t *testing.T) {
	e := httpexpect.New(t, testurl)
	contentType := "application/json"
	postdata := map[string]interface{}{
		"name":        "fred02",
		"description": "hellofred",
		"protocol":    "",
		"namespace":   "brobridge",
	}
	e.POST("/api/v1/routers").
		WithHeader("Content-Type", contentType).
		WithJSON(postdata).
		Expect().
		Status(http.StatusConflict).
		JSON().
		Object().
		ContainsKey("error")
}

func TestCreateRouterFail3(t *testing.T) {
	e := httpexpect.New(t, testurl)
	contentType := "application/json"
	postdata := map[string]interface{}{
		"name":        "fred02",
		"description": "hellofred",
		"protocol":    "http",
		"namespace":   "",
	}
	e.POST("/api/v1/routers").
		WithHeader("Content-Type", contentType).
		WithJSON(postdata).
		Expect().
		Status(http.StatusConflict).
		JSON().
		Object().
		ContainsKey("error")
}

func TestCreateRouterFail4(t *testing.T) {
	e := httpexpect.New(t, testurl)
	contentType := "application/json"
	postdata := map[string]interface{}{
		"name":        "fred01",
		"description": "hellofred",
		"protocol":    "http",
		"namespace":   "brobridge",
	}
	e.POST("/api/v1/routers").
		WithHeader("Content-Type", contentType).
		WithJSON(postdata).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("error")
}

// Get router testcase
func TestGetRouterPass1(t *testing.T) {
	e := httpexpect.New(t, testurl)

	obj := e.GET("/api/v1/routers").
		WithQuery("page", 1).
		WithQuery("limit", 10).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("meta").
		ContainsKey("routers")
	obj.Value("routers").Array().Length().Equal(6)
	obj.Value("meta").Object().Value("limit").Equal(10)
	obj.Value("meta").Object().Value("page").Equal(1)
}

func TestGetRouterPass2(t *testing.T) {
	e := httpexpect.New(t, testurl)

	obj := e.GET("/api/v1/routers").
		WithQuery("page", 1).
		WithQuery("limit", -1).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("meta").
		ContainsKey("routers")
	obj.Value("routers").Array().Length().Equal(6)
	obj.Value("meta").Object().Value("limit").Equal(-1)
	obj.Value("meta").Object().Value("page").Equal(1)
	obj.Value("meta").Object().Value("total").Equal(6)
}

func TestGetRouterPass3(t *testing.T) {
	e := httpexpect.New(t, testurl)

	obj := e.GET("/api/v1/routers").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("meta").
		ContainsKey("routers")
	obj.Value("routers").Array().Length().Equal(6)
	obj.Value("meta").Object().Value("limit").Equal(20)
	obj.Value("meta").Object().Value("page").Equal(1)
}

func TestGetRouterFail1(t *testing.T) {
	e := httpexpect.New(t, testurl)

	e.GET("/api/v1/routers").
		WithQuery("limit", -10).
		Expect().
		Status(http.StatusConflict).
		JSON().
		Object().
		ContainsKey("error")
}

// Enable router
func TestEnableRouterPass1(t *testing.T) {
	e := httpexpect.New(t, testurl)
	postdata := map[string]interface{}{
		"enabled": true,
	}

	obj := e.PUT("/api/v1/router/" + routers[0] + "/enabled").
		WithJSON(postdata).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("id")
	obj.Value("id").String().Equal(routers[0])

}
func TestEnableRouterPass2(t *testing.T) {
	e := httpexpect.New(t, testurl)
	postdata := map[string]interface{}{
		"enabled": false,
	}

	obj := e.PUT("/api/v1/router/" + routers[0] + "/enabled").
		WithJSON(postdata).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("id")
	obj.Value("id").String().Equal(routers[0])

}

func TestEnableRouterFail1(t *testing.T) {
	e := httpexpect.New(t, testurl)
	postdata := map[string]interface{}{
		"enabled": false,
	}

	e.PUT("/api/v1/router/" + "NoThisRouter" + "/enabled").
		WithJSON(postdata).
		Expect().
		Status(http.StatusConflict).
		JSON().
		Object().
		ContainsKey("error")
}
func TestEnableRouterFail2(t *testing.T) {
	e := httpexpect.New(t, testurl)

	e.PUT("/api/v1/router/" + routers[1] + "/enabled").
		Expect().
		Status(http.StatusBadRequest).
		JSON().
		Object().
		ContainsKey("error")
}

// Update router
func TestUpadteRouterPass1(t *testing.T) {
	e := httpexpect.New(t, testurl)
	postdata := map[string]interface{}{
		"name":        "XXXXXXX",
		"description": "xxxxxxx",
		"protocol":    "http",
		"namespace":   "xxxx",
	}

	obj := e.PUT("/api/v1/router/" + routers[0]).
		WithJSON(postdata).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("id")
	obj.Value("id").String().Equal(routers[0])
}

func TestUpadteRouterFail1(t *testing.T) {
	e := httpexpect.New(t, testurl)
	postdata := map[string]interface{}{
		"enabled": false,
	}

	e.PUT("/api/v1/router/" + "NoThisRouter").
		WithJSON(postdata).
		Expect().
		Status(http.StatusConflict).
		JSON().
		Object().
		ContainsKey("error")
}

func TestUpadteRouterFail2(t *testing.T) {
	e := httpexpect.New(t, testurl)
	postdata := map[string]interface{}{
		"name":        "",
		"description": "xxxxxxx",
		"protocol":    "",
		"namespace":   "xxxx",
	}

	e.PUT("/api/v1/router/" + routers[0]).
		WithJSON(postdata).
		Expect().
		Status(http.StatusConflict).
		JSON().
		Object().
		ContainsKey("error")
}

// Delete router
func TestDeleteRouterPass1(t *testing.T) {
	e := httpexpect.New(t, testurl)

	obj := e.DELETE("/api/v1/router/" + routers[0]).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("id")
	obj.Value("id").String().Equal(routers[0])

}
func TestDeleteRouterFail1(t *testing.T) {
	e := httpexpect.New(t, testurl)

	e.DELETE("/api/v1/router/" + "WillFail").
		Expect().
		Status(http.StatusConflict).
		JSON().
		Object().
		ContainsKey("error")
}

// Update router mapping
func TestUpdateRouterMappingPass1(t *testing.T) {
	e := httpexpect.New(t, testurl)
	// Create 2 gateway for mapping later

	gateway1 := map[string]interface{}{
		"name":        "fred01",
		"host":        "auth.brobridge.com",
		"description": "hellofred",
		"namespace":   "brobridge",
		"ports": []map[string]interface{}{
			{
				"port":     443,
				"protocol": "https",
				"cert":     "/etc/html/ooo.pem",
				"pkey":     "/etc/html/ooo.pem",
			},
		},
	}
	gateway2 := map[string]interface{}{
		"name":        "fred02",
		"host":        "auth.brobridge.com",
		"description": "hellofred",
		"namespace":   "brobridge",
		"ports": []map[string]interface{}{
			{
				"port":     443,
				"protocol": "https",
				"cert":     "/etc/html/ooo.pem",
				"pkey":     "/etc/html/ooo.pem",
			},
		},
	}
	gaid1 := e.POST("/api/v1/gateways").
		WithJSON(gateway1).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("id")
	gaid2 := e.POST("/api/v1/gateways").
		WithJSON(gateway2).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("id")
	fmt.Println("===================================================")
	willmaprouter1 := gaid1.Value("id").String().Raw()
	willmaprouter2 := gaid2.Value("id").String().Raw()
	fmt.Println("New gateway 1: " + willmaprouter1)
	fmt.Println("New gateway 2: " + willmaprouter2)
	fmt.Println("===================================================")

	body := map[string]interface{}{
		"gateways": []string{
			willmaprouter1,
			willmaprouter2,
		},
	}
	routerid := e.PUT("/api/v1/router/" + routers[1] + "/gateways").
		WithJSON(body).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("id")
	routerid.Value("id").String().Equal(routers[1])
	fmt.Println("===============================================================")
	fmt.Println("Mapped some gateways to this : " + routerid.Value("id").String().Raw())
	fmt.Println("===============================================================")
}

func TestUpdateRouterMappingFail1(t *testing.T) {
	e := httpexpect.New(t, testurl)

	e.PUT("/api/v1/router/" + "NoRouter" + "/gateways").
		Expect().
		Status(http.StatusConflict).
		JSON().
		Object().
		ContainsKey("error")
}

// Get router mapping
func TestGetRouterMappingPass1(t *testing.T) {
	e := httpexpect.New(t, testurl)

	e.GET("/api/v1/router/" + routers[1] + "/gateways").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("id").
		ContainsKey("gateways")
}

func TestGetRouterMappingFail1(t *testing.T) {
	e := httpexpect.New(t, testurl)

	e.GET("/api/v1/router/" + "NoRouter" + "/gateways").
		Expect().
		Status(http.StatusConflict).
		JSON().
		Object().
		ContainsKey("error")
}
