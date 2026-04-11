package main

import (
    "fmt"
    "net"
)

func main() {
    interfaces, err := net.Interfaces()
    if err != nil {
        fmt.Println("Erro ao listar interfaces:", err)
        return
    }

    fmt.Println("Listagem de Interfaces Ativas (IPv4):")
    fmt.Println("--------------------------------------")

    for _, iface := range interfaces {
        // Verifica se a interface está UP (ativa)
        if iface.Flags&net.FlagUp != 0 {
            addrs, err := iface.Addrs()
            if err != nil {
                continue
            }

            for _, addr := range addrs {
                // Verifica se o endereço é IPv4 (To4 != nil)
                // E remove o endereço de loopback (127.0.0.1) para limpar a lista
                ip, ok := addr.(*net.IPNet)
                if ok && ip.IP.To4() != nil && !ip.IP.IsLoopback() {
                    // Formata a saída: Nome da Interface [TAB] Endereço IP
                    fmt.Printf("%-15s %s\n", iface.Name, ip.IP.String())
                }
            }
        }
    }
    fmt.Println("--------------------------------------")
}
