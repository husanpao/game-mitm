//go:build darwin
// +build darwin

package gosysproxy

import (
	"errors"
	"os/exec"
	"strings"
)

// 获取所有网络服务名
func getNetworkServices() ([]string, error) {
	out, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	var services []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		// 第一行为 "An asterisk (*) denotes that a network service is disabled."
		if l == "" || strings.HasPrefix(l, "An asterisk") {
			continue
		}
		services = append(services, l)
	}
	return services, nil
}

// SetGlobalProxy 设置全局代理（HTTP/HTTPS）
func SetGlobalProxy(proxyServer string, bypasses ...string) error {
	if proxyServer == "" {
		return errors.New("代理服务器(proxyServer)配置为空")
	}
	hostPort := strings.Split(proxyServer, ":")
	if len(hostPort) != 2 {
		return errors.New("代理格式需为 host:port")
	}
	host, port := hostPort[0], hostPort[1]

	services, err := getNetworkServices()
	if err != nil {
		return err
	}
	for _, service := range services {
		// HTTP
		cmd := exec.Command("networksetup", "-setwebproxy", service, host, port)
		if err := cmd.Run(); err != nil {
			return err
		}
		cmd = exec.Command("networksetup", "-setwebproxystate", service, "on")
		if err := cmd.Run(); err != nil {
			return err
		}
		// HTTPS
		cmd = exec.Command("networksetup", "-setsecurewebproxy", service, host, port)
		if err := cmd.Run(); err != nil {
			return err
		}
		cmd = exec.Command("networksetup", "-setsecurewebproxystate", service, "on")
		if err := cmd.Run(); err != nil {
			return err
		}
		// 处理 bypass
		if len(bypasses) > 0 {
			bypassList := strings.Join(bypasses, ",")
			cmd = exec.Command("networksetup", "-setproxybypassdomains", service, bypassList)
			cmd.Run() // 不强制失败
		}
	}
	return nil
}

// Off 关闭所有代理
func Off() error {
	services, err := getNetworkServices()
	if err != nil {
		return err
	}
	for _, service := range services {
		exec.Command("networksetup", "-setwebproxystate", service, "off").Run()
		exec.Command("networksetup", "-setsecurewebproxystate", service, "off").Run()
	}
	return nil
}

// SetPAC 设置 PAC
func SetPAC(scriptLoc string) error {
	if scriptLoc == "" {
		return errors.New("PAC脚本地址(scriptLoc)配置为空")
	}
	services, err := getNetworkServices()
	if err != nil {
		return err
	}
	for _, service := range services {
		cmd := exec.Command("networksetup", "-setautoproxyurl", service, scriptLoc)
		if err := cmd.Run(); err != nil {
			return err
		}
		cmd = exec.Command("networksetup", "-setautoproxystate", service, "on")
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

// Status 获取状态（简单返回占位）
func Status() (*ProxyStatus, error) {
	return nil, errors.New("not implemented on darwin")
}

// Flush 无需实现，macOS 不用刷新
func Flush() error {
	return nil
}
