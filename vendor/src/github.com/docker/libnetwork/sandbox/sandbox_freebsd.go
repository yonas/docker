package sandbox

import (
  "sync"
  "net"

  "github.com/Sirupsen/logrus"
  "github.com/docker/libnetwork/types"
)


// NewSandbox provides a new sandbox instance created in an os specific way
// provided a key which uniquely identifies the sandbox
func NewSandbox(key string, osCreate bool) (Sandbox, error) {
  interfaces := []*Interface{}
  info := &Info{Interfaces: interfaces}

  return &FreebsdSandbox{path: key, sinfo: info}, nil
}

// GenerateKey generates a sandbox key based on the passed
// container id.
func GenerateKey(containerID string) string {
  maxLen := 12
  if len(containerID) < maxLen {
    maxLen = len(containerID)
  }

  return "net/" + containerID[:maxLen]
}


// Sandbox represents a network sandbox, identified by a specific key.  It
// holds a list of Interfaces, routes etc, and more can be added dynamically.
type FreebsdSandbox struct {
  path        string
  sinfo       *Info
  
  sync.Mutex
}

// The path where the network namespace is mounted.
func (n *FreebsdSandbox) Key() string {
  return n.path;
}

  // The collection of Interface previously added with the AddInterface
  // method. Note that this doesn't incude network interfaces added in any
  // other way (such as the default loopback interface which are automatically
  // created on creation of a sandbox).
func (n *FreebsdSandbox) Interfaces() []*Interface {
  return n.sinfo.Interfaces
}

  // Add an existing Interface to this sandbox. The operation will rename
  // from the Interface SrcName to DstName as it moves, and reconfigure the
  // interface according to the specified settings. The caller is expected
  // to only provide a prefix for DstName. The AddInterface api will auto-generate
  // an appropriate suffix for the DstName to disambiguate.
func (n *FreebsdSandbox) AddInterface(i *Interface) error {
  logrus.Debugf("[sandbox] add if");

  n.Lock()
  n.sinfo.Interfaces = append(n.sinfo.Interfaces, i)
  n.Unlock()

  return nil
}

  // Remove an interface from the sandbox by renaming to original name
  // and moving it out of the sandbox.
func (n *FreebsdSandbox) RemoveInterface(i *Interface) error {
  n.Lock()
  for index, intf := range n.sinfo.Interfaces {
    if intf == i {
      n.sinfo.Interfaces = append(n.sinfo.Interfaces[:index], n.sinfo.Interfaces[index+1:]...)
      break
    }
  }
  n.Unlock()

  return nil
}

  // Set default IPv4 gateway for the sandbox
func (n *FreebsdSandbox) SetGateway(gw net.IP) error {
  return nil
}

  // Set default IPv6 gateway for the sandbox
func (n *FreebsdSandbox) SetGatewayIPv6(gw net.IP) error {
  return nil
}

  // Add a static route to the sandbox.
func (n *FreebsdSandbox) AddStaticRoute(sr *types.StaticRoute) error {
  return nil
}

  // Remove a static route from the sandbox.
func (n *FreebsdSandbox) RemoveStaticRoute(sr *types.StaticRoute) error {
  return nil
}

  // Destroy the sandbox
func (n *FreebsdSandbox) Destroy() error {
  return nil
}