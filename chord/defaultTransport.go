package chord

import (
	"fmt"
)

// BlackholeTransport is used to provide an implemenation of the Transport that
// does not actually do anything. Any operation will result in an error.
type BlackholeTransport struct {
}

func (*BlackholeTransport) ListVnodes(host string) ([]*Vnode, error) {
	return nil, fmt.Errorf("Failed to connect! Blackhole: %s.", host)
}

func (*BlackholeTransport) Ping(vn *Vnode) (bool, error) {
	return false, nil
}

func (*BlackholeTransport) GetPredecessor(vn *Vnode) (*Vnode, error) {
	return nil, fmt.Errorf("Failed to connect! Blackhole: %s.", vn.String())
}

func (*BlackholeTransport) Notify(vn, self *Vnode) ([]*Vnode, error) {
	return nil, fmt.Errorf("Failed to connect! Blackhole: %s", vn.String())
}

func (*BlackholeTransport) FindSuccessors(vn *Vnode, n int, key []byte) ([]*Vnode, error) {
	return nil, fmt.Errorf("Failed to connect! Blackhole: %s", vn.String())
}

func (*BlackholeTransport) ClearPredecessor(target, self *Vnode) error {
	return fmt.Errorf("Failed to connect! Blackhole: %s", target.String())
}

func (*BlackholeTransport) SkipSuccessor(target, self *Vnode) error {
	return fmt.Errorf("Failed to connect! Blackhole: %s", target.String())
}

func (*BlackholeTransport) Register(v *Vnode, o VnodeRPC) {
}

func (*BlackholeTransport) Get(v *Vnode, key string) (string, error) {
	return "", fmt.Errorf("Failed to connect! Blackhole: %s", v.String())
}

func (*BlackholeTransport) Set(v *Vnode, key string, value string) error {
	return fmt.Errorf("Failed to connect! Blackhole: %s", v.String())
}

func (*BlackholeTransport) Delete(v *Vnode, key string) error {
	return fmt.Errorf("Failed to connect! Blackhole: %s", v.String())
}
