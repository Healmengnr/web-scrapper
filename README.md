# Go Web Scraper

Go dilinde yazilmis, web sitelerinin HTML icerigini ceken, ekran goruntusu alan ve sayfadaki tum URL'leri listeleyen bir CTI (Siber Tehdit Istihbarati) aracidir.

## Ozellikler

- Hedef sayfanin tam HTML icerigini indirir
- Sayfanin ekran goruntusunu PNG formatinda kaydeder
- Sayfadaki tum linkleri tespit edip listeler
- Baglanti hatalarini duzgun sekilde yakalar ve raporlar

## Gereksinimler

- Go 1.21 veya ustu
- Chrome veya Chromium tarayicisi

## Kurulum

```bash
# Repoyu klonla
git clone https://github.com/kullanici/web-scraper.git
cd web-scraper

# Bagimliliklari indir
go mod tidy
```

## Kullanim

```bash
# Dogrudan calistir
go run main.go https://yildizcti.com/

# veya derle ve calistir
go build -o web-scraper
./web-scraper https://yildizcti.com/
```

### Parametreler

| Parametre | Zorunlu | Varsayilan | Aciklama |
|-----------|---------|------------|----------|
| url | Evet | - | Taranacak web sitesinin URL'si |
| -output | Hayir | output | Cikti dizini |
| -timeout | Hayir | 60 | Zaman asimi (saniye) |

### Ornekler

```bash
# Basit kullanim
./web-scraper https://example.com

# Farkli cikti dizini
./web-scraper https://example.com -output ./veriler

# Uzun zaman asimi
./web-scraper https://yavas-site.com -timeout 120
```

## Cikti Dosyalari

Dosyalar domain adina gore klasorlenir. Subdomain varsa alt klasore kaydedilir. Ayni site tekrar taranirsa dosyalar numaralanir (2-screenshot.png gibi).

```
output/
├── example-com/
│   ├── page_content.html
│   ├── screenshot.png
│   ├── links.txt
│   └── www/              # www.example.com icin
│       └── ...
```

## Teknik Bilgiler

### Kullanilan Kutuphaneler

- **chromedp**: Headless Chrome kontrolu
- **goquery**: HTML parsing ve link cikarma

### Calisma Mantigi

1. Verilen URL'ye headless Chrome ile baglanir
2. Sayfa tamamen yuklendikten sonra HTML icerigini alir
3. Tam sayfa ekran goruntusu ceker
4. HTML'deki tum `<a>` etiketlerinden linkleri cikarir
5. Tum verileri dosyalara kaydeder

## Hata Yonetimi

Program asagidaki durumlarda hata mesaji verir:

- Gecersiz URL formati
- Baglanti zaman asimi
- Sayfa bulunamadi (404)
- DNS cozumleme hatasi
- SSL sertifika hatalari
