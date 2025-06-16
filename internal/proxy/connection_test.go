package proxy

import "testing"

func TestConnectionPoolAddAndRemove(t *testing.T) {
	p := &ConnectionPool{Users: make(map[string]*UserConnections)}
	conn := p.AddConnection("user", nil)
	if len(p.Users["user"].Connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(p.Users["user"].Connections))
	}
	p.RemoveConnection("user", conn.Conn)
	if _, ok := p.Users["user"]; ok {
		t.Fatalf("expected user removed")
	}
}

func TestConnectionPoolGetConnection(t *testing.T) {
	p := &ConnectionPool{Users: make(map[string]*UserConnections)}
	p.AddConnection("user", nil)
	c, err := p.GetConnection("user")
	if err != nil || c == nil {
		t.Fatalf("expected connection, got %v", err)
	}
}
