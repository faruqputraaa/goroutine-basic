package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// Struct untuk Menu Item
type MenuWarung struct {
	Nama   string
	Harga  float64
	Jumlah int
}

// Interface untuk Pesanan
type Pesanan interface {
	TotalHarga() float64
}

// Method untuk menghitung total harga pada MenuItem
func (m *MenuWarung) TotalHarga() float64 {
	return m.Harga * float64(m.Jumlah)
}

// Fungsi untuk validasi input angka bulat menggunakan regexp
func validasiInput(input string) error {
	if !regexp.MustCompile(`^\d+$`).MatchString(input) {
		return errors.New("input tidak valid, hanya menerima angka bulat.")
	}
	return nil
}

// Fungsi untuk menginput angka bulat dengan validasi
func inputAngka(prompt string) int {
	var input string
	for {
		fmt.Print(prompt)
		fmt.Scanln(&input)

		// Validasi input
		if err := validasiInput(input); err != nil {
			fmt.Println(err)
			fmt.Println("Silakan masukkan kembali.")
			continue
		}

		// Jika valid, konversi ke int dan return
		angka, _ := strconv.Atoi(input)
		return angka
	}
}

// Fungsi untuk menginput angka dengan validasi float64
func inputAngkaFloat(prompt string) float64 {
	var input string
	for {
		fmt.Print(prompt)
		fmt.Scanln(&input)

		// Validasi input
		if err := validasiInput(input); err != nil {
			fmt.Println(err)
			fmt.Println("Silakan masukkan kembali.")
			continue
		}

		angka, _ := strconv.ParseFloat(input, 64)
		return angka
	}
}

// Fungsi untuk menambah pesanan ke dalam map
func tambahPesanan(menu []MenuWarung, pesanan map[string]Pesanan) error {
	for {
		// Menampilkan daftar menu dan jumlah yang sudah dipesan
		fmt.Println("Daftar Menu:")
		for i, item := range menu {
			fmt.Printf("%d. %s - Rp%.2f (Pesanan: %d)\n", i+1, item.Nama, item.Harga, item.Jumlah)
		}

		// Memilih menu
		var pilihan int
		fmt.Print("Pilih menu (0 untuk selesai): ")
		_, err := fmt.Scanln(&pilihan)
		if err != nil || pilihan < 0 || pilihan > len(menu) {
			fmt.Println("Pilihan tidak valid, silakan coba lagi.")
			continue
		}

		// Jika pengguna memilih 0, keluar dari loop
		if pilihan == 0 {
			break
		}

		itemDipilih := &menu[pilihan-1]

		// Memilih jumlah pesanan (input dengan validasi)
		itemDipilih.Jumlah += inputAngka("Masukkan jumlah pesanan: ")

		// Memasukkan pesanan ke map
		pesanan[itemDipilih.Nama] = itemDipilih
		fmt.Printf("Anda telah memesan %d %s.\n", itemDipilih.Jumlah, itemDipilih.Nama)
	}
	return nil
}

// Fungsi untuk encode base64 pesanan
func encodePesanan(pesanan map[string]Pesanan) string {
	pesananStr := fmt.Sprintf("Pesanan: %+v", pesanan)
	return base64.StdEncoding.EncodeToString([]byte(pesananStr))
}

// Fungsi untuk menghitung kembalian
func hitungKembalian(totalHarga float64) float64 {
	var uangBayar float64
	for {
		uangBayar = inputAngkaFloat("Masukkan jumlah uang yang dibayarkan: ")

		if uangBayar >= totalHarga {
			return uangBayar - totalHarga
		} else {
			fmt.Println("Uang yang dibayarkan kurang, silakan masukkan kembali.")
		}
	}
}

// Fungsi untuk memproses pesanan dengan goroutine dan channel
func prosesPesanan(item *MenuWarung, ch chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Mengatur jeda 2 detik per item yang dipesan
	jeda := time.Duration(item.Jumlah) * 2 * time.Second

	ch <- fmt.Sprintf("Memproses pesanan untuk %d %s ...", item.Jumlah, item.Nama)
	time.Sleep(jeda)
	ch <- fmt.Sprintf("Pesanan %d %s selesai diproses dalam waktu %v detik.", item.Jumlah, item.Nama, jeda.Seconds())
}

// Fungsi utama untuk memproses pemesanan dan pembayaran
func prosesOrder(menu []MenuWarung) {
	pesanan := make(map[string]Pesanan)

	tambahPesanan(menu, pesanan)

	// Menampilkan pesanan dan total harga
	fmt.Println("Pesanan Anda:")
	totalHarga := 0.0
	for _, item := range pesanan {
		fmt.Printf("%s x%d = Rp%.2f\n", item.(*MenuWarung).Nama, item.(*MenuWarung).Jumlah, item.TotalHarga())
		totalHarga += item.TotalHarga()
	}
	totalBayar := totalHarga
	fmt.Printf("Total Bayar = Rp%.2f\n", totalBayar)

	// Encode pesanan ke base64
	encodedPesanan := encodePesanan(pesanan)
	fmt.Println("Pesanan terenkripsi (Base64):", encodedPesanan)

	// Menghitung kembalian
	kembalian := hitungKembalian(totalHarga)
	fmt.Printf("Uang Kembalian: Rp%.2f\n", kembalian)

	// Pemrosesan pesanan menggunakan goroutine
	ch := make(chan string, len(pesanan))
	var wg sync.WaitGroup
	wg.Add(len(pesanan))

	// Menjalankan goroutine untuk memproses setiap pesanan
	for _, item := range pesanan {
		go prosesPesanan(item.(*MenuWarung), ch, &wg)
	}

	// Menggunakan select untuk menangani channel
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Menampilkan status pemrosesan pesanan
	for {
		select {
		case status, ok := <-ch:
			if !ok {
				fmt.Println("Semua pesanan telah diproses.")
				return
			}
			fmt.Println(status)
		case <-time.After(5 * time.Second):
			fmt.Println("Maaf agak lama!")
			return
		}
	}
}

func main() {
	// Inisialisasi menu
	menu := []MenuWarung{
		{"Nasi Goreng", 20000, 0},
		{"Mie Goreng", 18000, 0},
		{"Ayam Bakar", 25000, 0},
	}
	defer fmt.Println("Program selesai!")

	// Menjalankan proses pemesanan dan pembayaran
	prosesOrder(menu)
}
