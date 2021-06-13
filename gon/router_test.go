package gon

import (
	"fmt"
	"reflect"
	"testing"
)

func newTestRouter() *router {
	r := newRouter()
	r.addRoute("GET", "/", nil)
	r.addRoute("GET", "/hello/:name", nil)
	r.addRoute("GET", "/hello/b/c", nil)
	r.addRoute("GET", "/hi/:name", nil)
	r.addRoute("GET", "/assets/*filepath", nil)
	return r
}

func TestParsePath(t *testing.T) {
	ok := reflect.DeepEqual(parsePath("/p/:name"), []string{"p", ":name"})
	ok = ok && reflect.DeepEqual(parsePath("/p/*"), []string{"p", "*"})
	ok = ok && reflect.DeepEqual(parsePath("/p/*name/*"), []string{"p", "*name"})
	if !ok {
		t.Fatal("test parsePath failed")
	}
}

func TestGetRoute(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRoute("GET", "/hello/gonzo")
	n1, ps1 := r.getRoute("GET", "/assets/file1.txt")

	if n == nil {
		t.Fatal("nil shouldn't be returned")
	}

	if n.path != "/hello/:name" {
		t.Fatal("should match /hello/:name")
	}

	if ps["name"] != "gonzo" {
		t.Fatal("name should be equal to 'gonzo'")
	}

	ok1 := n1.path == "/assets/*filepath" && ps1["filepath"] == "file1.txt"
	if !ok1 {
		t.Fatal("path should be /assets/*filepath & filepath should be file1.txt")
	}

	n2, ps2 := r.getRoute("GET", "/assets/css/test.css")
	ok2 := n2.path == "/assets/*filepath" && ps2["filepath"] == "css/test.css"
	if !ok2 {
		t.Fatal("path should be /assets/*filepath & filepath should be css/test.css")
	}

}

func TestGetRoutes(t *testing.T) {
	r := newTestRouter()
	nodes := r.getRoutes("GET")
	for i, n := range nodes {
		fmt.Println(i+1, n)
	}

	if len(nodes) != 5 {
		t.Fatal("the number of routes shoule be 4")
	}
}
