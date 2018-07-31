package frontend

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

//Route struct to easily add new roots
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

//NewRouter changed mux.Router func to work with the Rout struct
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		router.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	return router
}

//MiddleWare check the IP of client with the list of IPs in conf.json
func MiddleWare(h http.Handler) http.Handler {

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		addr := GetClientIPs(r)
		IPConf := GetConf().IP
		KeyConf := GetConf().Key

		if KeyCheck(r, KeyConf) == 1 {
			http.Error(rw, "Unauthorized API key", 401)
			return
		}

		for _, ipConf := range IPConf {
			for _, ipClient := range addr {
				if ipConf == ipClient {
					h.ServeHTTP(rw, r)
					return
				}
			}
		}
		http.Error(rw, "Unauthorized IP address", 401)
		return
	})
}

//GetClientIPs returns the IPs address of client
func GetClientIPs(r *http.Request) []string {
	//If X-FORWARDED-FOR structure is respected (first IP is the client's private IP address)
	//and separate with ", "
	//Header.Get will have all other IPs but not the last one used (last proxy or client if empty)
	var IPs []string
	if allIP := r.Header.Get("W-FORWARDED-FOR"); len(allIP) > 0 {
		IPs = strings.Split(allIP, ", ")
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	IPs = append(IPs, ip)
	return IPs
}

//KeyCheck checkst the API key given by the client in the Authorization header
func KeyCheck(r *http.Request, keys []string) int {
	fmt.Println(keys)
	fmt.Println("Header")
	for name, values := range r.Header {
		fmt.Println(name)
		fmt.Println(values)
		if name == "Authorization" {
			for _, key := range keys {
				if values[0] == key {
					return 0
				}
			}
		}
		break
	}
	return 1
}

//List of all possible routes
var routes = []Route{
	Route{
		"Echo",
		"POST",
		"/api/echo",
		Echo,
	},
	Route{
		"Task",
		"POST",
		"/api/task",
		Task,
	},
	Route{
		"Execute",
		"POST",
		"/api/execute",
		Execute,
	},
}
