package conn

func GetClient(key string) *Client {
	return clients[key]
}

var clients = make(map[string]*Client, 16)

func registerClient(c *Client) {
	clients[c.key] = c
}

func unregisterClient(c *Client) {
	delete(clients, c.key)
}
