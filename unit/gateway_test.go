//go:build integration

/*
*********************************************
Pilotwave API Unit Test
Test api url: /api/v1/gateway

Test function description:
1.  TestCreateGatewayPass: create ten gateway by pass
2.  TestCreateGatewayFail1: create gateway fail. (name empty)
3.  TestCreateGatewayFail2: create gateway fail. (host empty)
4.  TestGetGatewaysPass1: get gateways pass with default query
5.  TestGetGatewaysPass2: get gateways pass with query page=1 and limit=10
6.  TestGetGatewaysPass3: get gateways pass with query namespce=default
7.  TestGetGatewaysPass4: get gateways pass with query isDisabled=true
8.  TestGetGatewaysPass5: get gateways pass with query search=admin
9.  TestGetGatewaysPass6: get gateways pass with query search=admin, namespace=xxxx, isDisabled=true, limit=2,and page=1
10. TestGetGatewaysPass7: get gateways return data is empty with query namespace=nodata, limit=2andpage=1
11. TestGetGatewayPass1: get gateway pass.
12. TestGetGatewayFail1: get gateway fail, id is not uuid
13. TestGetGatewayFail2: get gateway fail, id is not found
14. TestEnableGatewayPass1: enable gateway pass
15. TestDeleteGatewayPass1: delete gateway pass
16. TestDeleteGatewayFail1: delete gateway fail, id is not uuid
17. TestUpdateGatewayPass1: update gateway pass
18. TestUpdateGatewayFail1: update gateway fail, name is empty
19. TestUpdateGatewayFail2: update gateway fail, host is empty
20. TestUpdateGatewayFail3: update gateway fail, namespace is empty
21. TestCreateRouterPass: create two router pass
22. TestUpdateGatewayMappingPass1: update gateway mapping two router pass
23. TestUpdateGatewayMappingPass2: update gateway clear router pass
24. TestUpdateGatewayMappingFail1: update gateway mapping fail, gateway id is not uuid
25. TestGetGatewayMappingPass1: get gateway mapping pass
26. TestGetGatewayMappingFail1: get gateway mapping fail, gateway id is empty
27. TestGetGatewayMappingFail2: get gateway mapping fail, gateway id is not uuid
28. TestGetGatewayMappingFail3: get gateway mapping fail, gateway id is not exist
29. TestCreateBlackWhiteListPass1: create gateway white list pass
30. TestCreateBlackWhiteListPass2: create gateway black list pass
31. TestCreateBlackWhiteListFail1: create gateway black and white list fail, gateway id is empty
32. TestCreateBlackWhiteListFail2: create gateway black and white list fail, gateway id is empty
33. TestCreateBlackWhiteListFail3: create gateway black and white list fail, gateway id is not exist
34. TestCreateBlackWhiteListFail4: create gateway black and white list fail, category not valid value
35. TestCreateBlackWhiteListFail5: create gateway black and white list fail, category is empty
36. TestCreateBlackWhiteListFail6: create gateway black and white list fail, domain is empty
37. TestGetBlackWhiteListPass1: get gateway black and white list pass, default query
38. TestGetBlackWhiteListPass2: get gateway black and white list pass, query is page=1 and limit=2
39. TestGetBlackWhiteListPass3: get gateway black and white list pass, query is category=blacklist
40. TestDeleteBlackWhiteListPass1: delete gateway black and white list pass
41. TestDeleteBlackWhiteListFail1: delete gateway black and white list fail, bwlist id is not uuid
**********************************************
*/
package test

import (
	"github.com/gavv/httpexpect"
	"net/http"
	"testing"
	//"log"
	"encoding/json"
)

// test url
var url string = "http://127.0.0.1:22112/api/v1"

// save gateway id
var gateway_id []string

/*
var gateway_id = []string{
	"84cdbc05-9de2-4fdb-9982-ebd81ac5a1be",
}
*/

// save router id
var router_id []string

/*
var router_id = []string{
	"31fe17d3-421d-4d3c-acf3-140bcaee563d",
	"e41f0659-b2ba-4c92-9af6-37f33b274726",
}
*/

// save bwlist id
var bwlist_id []string

/*
var bwlist_id = []string{
	"4f67cbd5-7170-455a-a235-d9bd50016e24",
}
*/

// Test create ten gateway success.
func TestCreateGatewayPass(t *testing.T) {
	e := httpexpect.New(t, url)
	for i := 0; i < 10; i++ {
		var postdata map[string]interface{}
		json_body := []byte(`
			{
				"name":"ga6",
				"host":"auth.brobridge.com",
				"description":"xxxxxxx",
				"ports":[
				   {
					  "port":443,
					  "protocol":"https",
					  "cert":"/etc/html/ooo.pem",
					  "pkey":"/etc/html/ooo.pem"
				   },
				   {
					  "port":8080,
					  "protocol":"http",
					  "cert":"",
					  "pkey":""
				   }
				],
				"namespace":"test"
			}
		`)
		json.Unmarshal(json_body, &postdata)
		contentType := "application/json;charset=utf-8"
		obj := e.POST("/gateways").WithHeader("ContentType", contentType).WithJSON(postdata).
			Expect().Status(http.StatusOK).JSON().Object().
			ContainsKey("id")
		gateway_id = append(gateway_id, obj.Value("id").String().Raw())
	}
}

// Test create gateway fail (name empty).
func TestCreateGatewayFail1(t *testing.T) {
	e := httpexpect.New(t, url)
	var postdata map[string]interface{}
	json_body := []byte(`
		{
			"name":"",
			"host":"auth.brobridge.com",
			"description":"xxxxxxx",
			"ports":[
			   {
				  "port":443,
				  "protocol":"https",
				  "cert":"/etc/html/ooo.pem",
				  "pkey":"/etc/html/ooo.pem"
			   },
			   {
				  "port":8080,
				  "protocol":"http",
				  "cert":"",
				  "pkey":""
			   }
			],
			"namespace":"test"
		}
	`)
	json.Unmarshal(json_body, &postdata)
	contentType := "application/json;charset=utf-8"
	e.POST("/gateways").WithHeader("ContentType", contentType).WithJSON(postdata).
		Expect().Status(http.StatusConflict).JSON().Object().
		ValueEqual("error", "name: cannot be blank.")
}

// Test create gateway fail (host empty).
func TestCreateGatewayFail2(t *testing.T) {
	e := httpexpect.New(t, url)
	var postdata map[string]interface{}
	json_body := []byte(`
		{
			"name":"ga6",
			"host":"",
			"description":"xxxxxxx",
			"ports":[
			   {
				  "port":443,
				  "protocol":"https",
				  "cert":"/etc/html/ooo.pem",
				  "pkey":"/etc/html/ooo.pem"
			   },
			   {
				  "port":8080,
				  "protocol":"http",
				  "cert":"",
				  "pkey":""
			   }
			],
			"namespace":"test"
		}
	`)
	json.Unmarshal(json_body, &postdata)
	contentType := "application/json;charset=utf-8"
	e.POST("/gateways").WithHeader("ContentType", contentType).WithJSON(postdata).
		Expect().Status(http.StatusConflict).JSON().Object().
		ValueEqual("error", "host: cannot be blank.")
}

// Test get gateways with defualt query.
func TestGetGatewaysPass1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.GET("/gateways").WithHeader("ContentType", contentType).
		Expect().Status(http.StatusOK).JSON().Object()
	obj.Keys().ContainsOnly("meta", "gateways")
	obj.ContainsMap(map[string]interface{}{ // success
		"meta": map[string]interface{}{
			"page":  1,
			"limit": 20,
		},
	})
}

// Test get gateways. query is page=1 and limit=10.
func TestGetGatewaysPass2(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.GET("/gateways").WithHeader("ContentType", contentType).
		WithQuery("page", 1).WithQuery("limit", 10).
		Expect().Status(http.StatusOK).JSON().Object()
	obj.Keys().ContainsOnly("meta", "gateways")
	obj.ContainsMap(map[string]interface{}{
		"meta": map[string]interface{}{
			"page":  1,
			"limit": 10,
		},
	})
}

// Test get gateways. query is namespace=default.
func TestGetGatewaysPass3(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.GET("/gateways").WithHeader("ContentType", contentType).
		WithQuery("namespace", "default").
		Expect().Status(http.StatusOK).JSON().Object()
	obj.Keys().ContainsOnly("meta", "gateways")
	array := obj.Value("gateways").Array()
	for _, val := range array.Iter() {
		val.Object().ValueEqual("namespace", "default")
	}
}

// Test get gateways. query is isDisabled=true.
func TestGetGatewaysPass4(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.GET("/gateways").WithHeader("ContentType", contentType).
		WithQuery("isDisabled", true).
		Expect().Status(http.StatusOK).JSON().Object()
	obj.Keys().ContainsOnly("meta", "gateways")
	array := obj.Value("gateways").Array()
	for _, val := range array.Iter() {
		val.Object().ValueEqual("isDisabled", true)
	}
}

// Test get gateways. query is search=admin.
func TestGetGatewaysPass5(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.GET("/gateways").WithHeader("ContentType", contentType).
		WithQuery("search", "admin").
		Expect().Status(http.StatusOK).JSON().Object()
	obj.Keys().ContainsOnly("meta", "gateways")
	array := obj.Value("gateways").Array()
	for _, val := range array.Iter() {
		val.Object().ValueEqual("name", "admin")
	}
}

// Test get gateways. query is search=admin&namespace=xxxx&isDisabled=true&limit=2&page=1
func TestGetGatewaysPass6(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.GET("/gateways").WithHeader("ContentType", contentType).
		WithQuery("search", "admin").WithQuery("namespace", "xxxx").
		WithQuery("isDisabled", true).WithQuery("limit", 2).
		WithQuery("page", 1).
		Expect().Status(http.StatusOK).JSON().Object()
	obj.Keys().ContainsOnly("meta", "gateways")
	obj.ContainsMap(map[string]interface{}{
		"meta": map[string]interface{}{
			"page":  1,
			"limit": 2,
		},
	})
	array := obj.Value("gateways").Array()
	for _, val := range array.Iter() {
		val.Object().ContainsMap(map[string]interface{}{
			"name":       "admin",
			"namespace":  "xxxx",
			"isDisabled": true,
		})
	}
}

// Test get gateways. query is namespace=nodata&limit=2&page=1
func TestGetGatewaysPass7(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.GET("/gateways").WithHeader("ContentType", contentType).
		WithQuery("namespace", "nodata").WithQuery("limit", 2).WithQuery("page", 1).
		Expect().Status(http.StatusOK).JSON().Object()
	obj.Keys().ContainsOnly("meta", "gateways")
	obj.ContainsMap(map[string]interface{}{
		"meta": map[string]interface{}{
			"page":  1,
			"limit": 2,
		},
	})
	array := obj.Value("gateways").Array()
	array.Empty()
}

// Test get gateway.
func TestGetGatewayPass1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.GET("/gateway/"+gateway_id[0]).WithHeader("ContentType", contentType).
		Expect().Status(http.StatusOK).JSON().Object()
	obj.ValueEqual("id", gateway_id[0])
}

// Test get gateway, id is not uuid.
func TestGetGatewayFail1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.GET("/gateway/123456").WithHeader("ContentType", contentType).
		Expect().Status(http.StatusConflict).JSON().Object()
	obj.ValueEqual("error", "must be a valid UUID")
}

// Test get gateway, id not found.
func TestGetGatewayFail2(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.GET("/gateway/0ab0f651-9f1d-4c06-8dca-c88f23202781").WithHeader("ContentType", contentType).
		Expect().Status(http.StatusInternalServerError).JSON().Object()
	obj.ValueEqual("error", "record not found")
}

// Test enable gateway. enabled = true
func TestEnableGatewayPass1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	var postdata map[string]interface{}
	json_body := []byte(`
	{ 
		"enabled": true
	}
	`)
	json.Unmarshal(json_body, &postdata)
	obj := e.PUT("/gateway/"+gateway_id[0]+"/enabled").WithHeader("ContentType", contentType).
		WithJSON(postdata).Expect().Status(http.StatusOK).JSON().Object()
	obj.ValueEqual("id", gateway_id[0])
}

// Test delete gateway.
func TestDeleteGatewayPass1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.DELETE("/gateway/"+gateway_id[0]).WithHeader("ContentType", contentType).
		Expect().Status(http.StatusOK).JSON().Object()
	obj.ValueEqual("id", gateway_id[0])
	gateway_id = gateway_id[1:len(gateway_id)]
}

// Test delete gateway. id is not uuid
func TestDeleteGatewayFail1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.DELETE("/gateway/123456").WithHeader("ContentType", contentType).
		Expect().Status(http.StatusConflict).JSON().Object()
	obj.ValueEqual("error", "must be a valid UUID")
}

// Test update gateway.
func TestUpdateGatewayPass1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	var postdata map[string]interface{}
	json_body := []byte(`
	{ 
		"name": "ga77",
		"host": "xxx.brobridge.com",
		"description": "desssss",
		"ports": [{
		  "port": 4433,
		  "protocol": "https",
		  "cert": "/etc/html/ooo.pem",
		  "pkey": "/etc/html/ooo.pem"
		},
		{
		  "port": 8080,
		  "protocol": "http",
		  "cert": "",
		  "pkey": ""
		}],
		"namespace": "test"
	}
	`)
	json.Unmarshal(json_body, &postdata)
	obj := e.PUT("/gateway/"+gateway_id[0]).WithHeader("ContentType", contentType).
		WithJSON(postdata).Expect().Status(http.StatusOK).JSON().Object()
	obj.ValueEqual("id", gateway_id[0])
}

// Test update gateway. name is empty
func TestUpdateGatewayFail1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	var postdata map[string]interface{}
	json_body := []byte(`
	{ 
		"name": "",
		"host": "xxx.brobridge.com",
		"description": "desssss",
		"ports": [{
		  "port": 4433,
		  "protocol": "https",
		  "cert": "/etc/html/ooo.pem",
		  "pkey": "/etc/html/ooo.pem"
		},
		{
		  "port": 8080,
		  "protocol": "http",
		  "cert": "",
		  "pkey": ""
		}],
		"namespace": "test"
	}
	`)
	json.Unmarshal(json_body, &postdata)
	obj := e.PUT("/gateway/"+gateway_id[0]).WithHeader("ContentType", contentType).
		WithJSON(postdata).Expect().Status(http.StatusConflict).JSON().Object()
	obj.ValueEqual("error", "name: cannot be blank.")
}

// Test update gateway. host is empty
func TestUpdateGatewayFail2(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	var postdata map[string]interface{}
	json_body := []byte(`
	{ 
		"name": "ga66",
		"host": "",
		"description": "desssss",
		"ports": [{
		  "port": 4433,
		  "protocol": "https",
		  "cert": "/etc/html/ooo.pem",
		  "pkey": "/etc/html/ooo.pem"
		},
		{
		  "port": 8080,
		  "protocol": "http",
		  "cert": "",
		  "pkey": ""
		}],
		"namespace": "test"
	}
	`)
	json.Unmarshal(json_body, &postdata)
	obj := e.PUT("/gateway/"+gateway_id[0]).WithHeader("ContentType", contentType).
		WithJSON(postdata).Expect().Status(http.StatusConflict).JSON().Object()
	obj.ValueEqual("error", "host: cannot be blank.")
}

// Test update gateway. namespace is empty
func TestUpdateGatewayFail3(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	var postdata map[string]interface{}
	json_body := []byte(`
	{ 
		"name": "ga66",
		"host": "xxx.brobridge.com",
		"description": "desssss",
		"ports": [{
		  "port": 4433,
		  "protocol": "https",
		  "cert": "/etc/html/ooo.pem",
		  "pkey": "/etc/html/ooo.pem"
		},
		{
		  "port": 8080,
		  "protocol": "http",
		  "cert": "",
		  "pkey": ""
		}],
		"namespace": ""
	}
	`)
	json.Unmarshal(json_body, &postdata)
	obj := e.PUT("/gateway/"+gateway_id[0]).WithHeader("ContentType", contentType).
		WithJSON(postdata).Expect().Status(http.StatusConflict).JSON().Object()
	obj.ValueEqual("error", "namespace: cannot be blank.")
}

// Test create two router success.
func TestCreateRouterPass(t *testing.T) {
	e := httpexpect.New(t, url)
	for i := 0; i < 2; i++ {
		var postdata map[string]interface{}
		json_body := []byte(`
			{
				"name":"test",
				"description":"xxxxxxx",
				"protocol":"http",
				"namespace":"xxxx"
			}
		`)
		json.Unmarshal(json_body, &postdata)
		contentType := "application/json;charset=utf-8"
		obj := e.POST("/routers").WithHeader("ContentType", contentType).WithJSON(postdata).
			Expect().Status(http.StatusOK).JSON().Object().
			ContainsKey("id")
		router_id = append(router_id, obj.Value("id").String().Raw())
	}
}

// Test update gateway mapping two router.
func TestUpdateGatewayMappingPass1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	// test
	obj := e.PUT("/gateway/"+gateway_id[0]+"/routers").WithHeader("ContentType", contentType).
		WithJSON(router_id).Expect().Status(http.StatusOK).JSON().Object()
	obj.ValueEqual("id", gateway_id[0])

	// check
	obj = e.GET("/gateway/"+gateway_id[0]+"/routers").WithHeader("ContentType", contentType).
		Expect().Status(http.StatusOK).JSON().Object()
	obj.Value("routers").Array()
	for n, val := range obj.Value("routers").Array().Iter() {
		val.Object().ValueEqual("id", router_id[n])
	}
}

// Test update gateway clear router
func TestUpdateGatewayMappingPass2(t *testing.T) {
	var router []string
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"

	// test
	obj := e.PUT("/gateway/"+gateway_id[0]+"/routers").WithHeader("ContentType", contentType).
		WithJSON(router).Expect().Status(http.StatusOK).JSON().Object()
	obj.ValueEqual("id", gateway_id[0])

	// check
	obj = e.GET("/gateway/"+gateway_id[0]+"/routers").WithHeader("ContentType", contentType).
		Expect().Status(http.StatusOK).JSON().Object()
	obj.ValueEqual("id", gateway_id[0])
	obj.Value("routers").Array().Empty()
}

// Test update gateway mapping. gateway id not is uuid
func TestUpdateGatewayMappingFail1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"

	obj := e.PUT("/gateway/123456/routers").WithHeader("ContentType", contentType).
		WithJSON(router_id).Expect().Status(http.StatusConflict).JSON().Object()
	obj.ValueEqual("error", "must be a valid UUID")
}

// Test get gateway mapping.
func TestGetGatewayMappingPass1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"

	obj := e.GET("/gateway/"+gateway_id[0]+"/routers").WithHeader("ContentType", contentType).
		Expect().Status(http.StatusOK).JSON().Object()
	obj.ValueEqual("id", gateway_id[0])
	obj.ContainsKey("routers")
}

// Test get gateway mapping. gateway id empty.
func TestGetGatewayMappingFail1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"

	obj := e.GET("/gateway//routers").WithHeader("ContentType", contentType).
		Expect().Status(http.StatusConflict).JSON().Object()
	obj.ValueEqual("error", "cannot be blank")
}

// Test get gateway mapping. gateway id not is uuid.
func TestGetGatewayMappingFail2(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"

	obj := e.GET("/gateway/123456/routers").WithHeader("ContentType", contentType).
		Expect().Status(http.StatusConflict).JSON().Object()
	obj.ValueEqual("error", "must be a valid UUID")
}

// Test get gateway mapping. gateway id not exist.
func TestGetGatewayMappingFail3(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"

	obj := e.GET("/gateway/ec3c94ec-1ac9-4ab3-a042-85f1bad94181/routers").WithHeader("ContentType", contentType).
		Expect().Status(http.StatusInternalServerError).JSON().Object()
	obj.ValueEqual("error", "record not found")
}

// Test create gateway white list.
func TestCreateBlackWhiteListPass1(t *testing.T) {
	e := httpexpect.New(t, url)
	var postdata map[string]interface{}
	json_body := []byte(`
		{
			"domain":"auth.brobridge.com",
			"description":"xxxxx",
			"category":"whitelist"
		}
	`)
	json.Unmarshal(json_body, &postdata)
	contentType := "application/json;charset=utf-8"
	obj := e.POST("/gateway/"+gateway_id[0]+"/bwlists").WithHeader("ContentType", contentType).
		WithJSON(postdata).Expect().Status(http.StatusOK).JSON().Object().
		ContainsKey("id")
	bwlist_id = append(bwlist_id, obj.Value("id").String().Raw())
}

// Test create gateway black list.
func TestCreateBlackWhiteListPass2(t *testing.T) {
	e := httpexpect.New(t, url)
	var postdata map[string]interface{}
	json_body := []byte(`
		{
			"domain":"test.brobridge.com",
			"description":"xxxxx",
			"category":"blacklist"
		}
	`)
	json.Unmarshal(json_body, &postdata)
	contentType := "application/json;charset=utf-8"
	obj := e.POST("/gateway/"+gateway_id[0]+"/bwlists").WithHeader("ContentType", contentType).
		WithJSON(postdata).Expect().Status(http.StatusOK).JSON().Object().
		ContainsKey("id")
	bwlist_id = append(bwlist_id, obj.Value("id").String().Raw())
}

// Test create black and white list. gateway id empty.
func TestCreateBlackWhiteListFail1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	var postdata map[string]interface{}
	json_body := []byte(`
		{
			"domain":"test.brobridge.com",
			"description":"xxxxx",
			"category":"blacklist"
		}
	`)
	json.Unmarshal(json_body, &postdata)
	obj := e.POST("/gateway//bwlists").WithHeader("ContentType", contentType).
		WithJSON(postdata).Expect().Status(http.StatusConflict).JSON().Object()
	obj.ValueEqual("error", "cannot be blank")
}

// Test create black and white list. gateway id not is uuid.
func TestCreateBlackWhiteListFail2(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	var postdata map[string]interface{}
	json_body := []byte(`
		{
			"domain":"test.brobridge.com",
			"description":"xxxxx",
			"category":"blacklist"
		}
	`)
	json.Unmarshal(json_body, &postdata)
	obj := e.POST("/gateway/123456/bwlists").WithHeader("ContentType", contentType).
		WithJSON(postdata).Expect().Status(http.StatusConflict).JSON().Object()
	obj.ValueEqual("error", "must be a valid UUID")
}

// Test create black and white list. gateway id not exist.
func TestCreateBlackWhiteListFail3(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	var postdata map[string]interface{}
	json_body := []byte(`
		{
			"domain":"test.brobridge.com",
			"description":"xxxxx",
			"category":"blacklist"
		}
	`)
	json.Unmarshal(json_body, &postdata)
	obj := e.POST("/gateway/ec3c94ec-1ac9-4ab3-a042-85f1bad94181/bwlists").WithHeader("ContentType", contentType).
		WithJSON(postdata).Expect().Status(http.StatusInternalServerError).JSON().Object()
	obj.ValueEqual("error", "record not found")
}

// Test create gateway black and white list. category not valid value.
func TestCreateBlackWhiteListFail4(t *testing.T) {
	e := httpexpect.New(t, url)
	var postdata map[string]interface{}
	json_body := []byte(`
		{
			"domain":"test.brobridge.com",
			"description":"xxxxx",
			"category":"test"
		}
	`)
	json.Unmarshal(json_body, &postdata)
	contentType := "application/json;charset=utf-8"
	e.POST("/gateway/"+gateway_id[0]+"/bwlists").WithHeader("ContentType", contentType).WithJSON(postdata).
		Expect().Status(http.StatusConflict).JSON().Object().
		ValueEqual("error", "category: must be a valid value.")
}

// Test create gateway black and white list. category empty.
func TestCreateBlackWhiteListFail5(t *testing.T) {
	e := httpexpect.New(t, url)
	var postdata map[string]interface{}
	json_body := []byte(`
		{
			"domain":"test.brobridge.com",
			"description":"xxxxx",
			"category":""
		}
	`)
	json.Unmarshal(json_body, &postdata)
	contentType := "application/json;charset=utf-8"
	e.POST("/gateway/"+gateway_id[0]+"/bwlists").WithHeader("ContentType", contentType).WithJSON(postdata).
		Expect().Status(http.StatusConflict).JSON().Object().
		ValueEqual("error", "category: cannot be blank.")
}

// Test create gateway black and white list. domain empty.
func TestCreateBlackWhiteListFail6(t *testing.T) {
	e := httpexpect.New(t, url)
	var postdata map[string]interface{}
	json_body := []byte(`
		{
			"domain":"",
			"description":"xxxxx",
			"category":"whitelist"
		}
	`)
	json.Unmarshal(json_body, &postdata)
	contentType := "application/json;charset=utf-8"
	e.POST("/gateway/"+gateway_id[0]+"/bwlists").WithHeader("ContentType", contentType).WithJSON(postdata).
		Expect().Status(http.StatusConflict).JSON().Object().
		ValueEqual("error", "domain: cannot be blank.")
}

// Test get gateway black and white list by default query.
func TestGetBlackWhiteListPass1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.GET("/gateway/"+gateway_id[0]+"/bwlists").WithHeader("ContentType", contentType).
		Expect().Status(http.StatusOK).JSON().Object()
	obj.Keys().ContainsOnly("meta", "lists")
	obj.ContainsMap(map[string]interface{}{ // success
		"meta": map[string]interface{}{
			"page":  1,
			"limit": 20,
		},
	})
}

// Test get gateway black and white list, query ?page=1&limit=2.
func TestGetBlackWhiteListPass2(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.GET("/gateway/"+gateway_id[0]+"/bwlists").WithHeader("ContentType", contentType).
		WithQuery("page", 1).WithQuery("limit", 2).
		Expect().Status(http.StatusOK).JSON().Object()
	obj.Keys().ContainsOnly("meta", "lists")
	obj.ContainsMap(map[string]interface{}{ // success
		"meta": map[string]interface{}{
			"page":  1,
			"limit": 2,
		},
	})
}

// Test get gateway black and white list, query ?category=blacklist.
func TestGetBlackWhiteListPass3(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.GET("/gateway/"+gateway_id[0]+"/bwlists").WithHeader("ContentType", contentType).
		WithQuery("category", "blacklist").
		Expect().Status(http.StatusOK).JSON().Object()
	obj.Keys().ContainsOnly("meta", "lists")
	array := obj.Value("lists").Array()
	for _, val := range array.Iter() {
		val.Object().ContainsMap(map[string]interface{}{
			"category": "blacklist",
		})
	}
}

// Test delete gateway black and white list,
func TestDeleteBlackWhiteListPass1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.DELETE("/bwlist/"+bwlist_id[0]).WithHeader("ContentType", contentType).
		Expect().Status(http.StatusOK).JSON().Object()
	obj.ValueEqual("id", bwlist_id[0])
	bwlist_id = bwlist_id[1:len(bwlist_id)]
}

// Test delete gateway black and white list, bwlist id is not uuid.
func TestDeleteBlackWhiteListFail1(t *testing.T) {
	e := httpexpect.New(t, url)
	contentType := "application/json;charset=utf-8"
	obj := e.DELETE("/bwlist/123456").WithHeader("ContentType", contentType).
		Expect().Status(http.StatusConflict).JSON().Object()
	obj.ValueEqual("error", "must be a valid UUID")
}
