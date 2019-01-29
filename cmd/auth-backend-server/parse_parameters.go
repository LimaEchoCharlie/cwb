package main

import (
	"fmt"
	"net/url"
)

const (
	// request parameters
	username     = "username"
	password     = "password"
	vhostName    = "vhost"
	clientIP     = "ip"
	resourceType = "resource"
	resourceName = "name"
	exchangeName = "name"
	accessLevel  = "permission"
	routingKey   = "routing_key"
)

// return the zeroth string in the value array with the supplied key. Returns an error if missing,
func getSingle(values url.Values, key string) (string, error) {
	list, ok := values[key]
	if !ok || len(list) == 0 {
		return "", fmt.Errorf("missing parameter %v", key)
	}
	return list[0], nil
}

// userAuthN holds the parameters contained in a user authentication request
type userAuthN struct {
	username, password string
}

func (i *userAuthN) String() string {
	return fmt.Sprintf("userAuthN {username: %s, password: %s}", i.username, i.password)
}

// Parse from url values into struct. Returns an error if an expected parameter is missing.
func (i *userAuthN) Parse(values url.Values) error {
	var err error
	if i.username, err = getSingle(values, username); err != nil {
		return err
	}
	if i.password, err = getSingle(values, password); err != nil {
		return err
	}
	return nil
}

// vHostAuthZ holds the parameters contained in a vhost authorisation request
type vHostAuthZ struct {
	username, vhostName, clientIP string
}

func (i *vHostAuthZ) String() string {
	return fmt.Sprintf("vHostAuthZ {username: %s, vhost: %s, client IP: %s}", i.username, i.vhostName, i.clientIP)
}

// Parse from url values into struct. Returns an error if an expected parameter is missing.
func (i *vHostAuthZ) Parse(values url.Values) error {
	var err error
	if i.username, err = getSingle(values, username); err != nil {
		return err
	}
	if i.vhostName, err = getSingle(values, vhostName); err != nil {
		return err
	}
	if i.clientIP, err = getSingle(values, clientIP); err != nil {
		return err
	}
	return nil
}

// resourceAuthZ holds the parameters contained in a resource authorisation request
type resourceAuthZ struct {
	username, vhostName, resourceType, resourceName, accessLevel string
}

func (i *resourceAuthZ) String() string {
	return fmt.Sprintf("resourceAuthZ {username: %s, vhost: %s, resource {type: %s, name: %s}, access level: %s}",
		i.username, i.vhostName, i.resourceType, i.resourceName, i.accessLevel)
}

// Parse from url values into struct. Returns an error if an expected parameter is missing.
func (i *resourceAuthZ) Parse(values url.Values) error {
	var err error
	if i.username, err = getSingle(values, username); err != nil {
		return err
	}
	if i.vhostName, err = getSingle(values, vhostName); err != nil {
		return err
	}
	if i.resourceType, err = getSingle(values, resourceType); err != nil {
		return err
	}
	if i.resourceName, err = getSingle(values, resourceName); err != nil {
		return err
	}
	if i.accessLevel, err = getSingle(values, accessLevel); err != nil {
		return err
	}
	return nil
}
