package main

import (
	"archive/tar"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// OVF-Struktur (vereinfacht)
type Envelope struct {
	XMLName       xml.Name      `xml:"Envelope"`
	VirtualSystem VirtualSystem `xml:"VirtualSystem"`
}

type ProductSection struct {
	Product    string `xml:"Product"`
	ProductUrl string `xml:"ProductUrl"`
	Vendor     string `xml:"Vendor"`
	VendorUrl  string `xml:"VendorUrl"`
	Version    string `xml:"Version"`
}

type VirtualSystem struct {
	ps ProductSection `xml:"ProductSection"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please give path to OVA")
	}
	ovaPath := os.Args[1]

	file, err := os.Open(ovaPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	tarReader := tar.NewReader(file)

	var ovfData []byte

	// Suche die .ovf-Datei im TAR
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if strings.HasSuffix(header.Name, ".ovf") {
			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, tarReader); err != nil {
				log.Fatal(err)
			}
			ovfData = buf.Bytes()
			break
		}
	}

	if ovfData == nil {
		log.Fatal("Keine OVF-Datei in der OVA gefunden")
	}

	// XML parsen
	var env Envelope
	if err := xml.Unmarshal(ovfData, &env); err != nil {
		log.Fatal(err)
	}

	// Ausgabe
	fmt.Println("Product:", env.VirtualSystem.ps.Product)
	fmt.Println("ProductUrl:", env.VirtualSystem.ps.ProductUrl)
	fmt.Println("Vendor:", env.VirtualSystem.ps.Vendor)
	fmt.Println("VendorUrl:", env.VirtualSystem.ps.VendorUrl)
	fmt.Println("Version:", env.VirtualSystem.ps.Version)
}
