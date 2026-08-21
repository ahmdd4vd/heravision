# Langkah Setelah MD1 MinArea

## Posisi sekarang

Kita sudah memiliki 30-image annotation fixture dan sudah membuktikan bahwa `MinArea=8` mengurangi proposal explosion. Namun fixture tersebut baru berupa daftar sample dan contact sheet; belum menjadi ground truth yang dapat menghitung precision/recall. Karena itu langkah berikutnya adalah memperbaiki **measurement layer**, bukan langsung menambah model.

## 1. Membuat ground truth region

Untuk 30 fixture, setiap gambar akan memiliki region visible yang dapat diaudit. Region harus berupa box atau polygon yang benar-benar terlihat, bukan label class yang dipaksakan. Contoh region adalah objek utama, panel dokumen, blok diagram, node, atau area foreground yang memiliki batas visual cukup jelas.

Schema minimalnya:

```json
{
  "sample_id": "blind-md0-imagenette-...",
  "regions": [
    {
      "id": "gt-001",
      "bbox": {"x": 0.12, "y": 0.18, "w": 0.44, "h": 0.51},
      "visibility": "visible",
      "source": "human-review"
    }
  ]
}
```

Untuk gambar yang batasnya ambigu, anotasi harus dapat memakai `uncertain` atau `ignore`, bukan memaksa keputusan palsu. Ini penting untuk mengukur abstention dan menghindari penalti terhadap engine karena annotation noise.

## 2. Membuat ground truth relation

Relation hanya memakai predicate yang terlihat langsung: `above`, `below`, `left_of`, `right_of`, `inside`, `overlapping`, dan `touching`. Relation semantik seperti `uses`, `belongs_to`, `is_a`, atau `causes` belum boleh dimasukkan karena B1 belum memiliki semantic evidence.

Setiap relation menyimpan evidence basis:

```json
{
  "from": "gt-001",
  "to": "gt-002",
  "predicate": "above",
  "visibility": "visible",
  "source": "human-review"
}
```

## 3. Menghitung baseline yang benar

Setelah anotasi tersedia, kita hitung B0, B1-old, dan B1-new dengan aturan sama. Metriknya adalah region precision, recall, F1 pada IoU 0,25/0,50/0,75, mean matched IoU, fragment rate, false relation rate, unsupported claim rate, dan abstention coverage.

B1-new hanya boleh dianggap lebih baik jika peningkatan tidak berasal dari satu domain saja. Minimal ada tabel per domain untuk Imagenette, Imagewoof, dan Wikimedia.

## 4. Failure-driven boundary merge

Failure gallery saat ini menunjukkan proposal terlalu banyak. Setelah ground truth, kita dapat membedakan dua kasus yang saat ini sama-sama terlihat sebagai banyak region:

| Kasus | Perbaikan yang dicari |
|---|---|
| Satu objek terpecah menjadi banyak region | Merge berdasarkan boundary lemah dan scale stability |
| Banyak objek menempel menjadi satu region | Boundary barrier dan local contrast |
| Background bertekstur dianggap objek | Texture uncertainty dan negative evidence |
| Diagram node/connector bercampur | Skeleton/line evidence dan spatial graph |

Perbaikan pertama yang paling ilmiah adalah menambah **scale stability**. Region yang muncul konsisten pada beberapa resolusi diberi bobot lebih tinggi; region yang hanya muncul di satu scale diperlakukan sebagai kandidat lemah. Ini lebih baik daripada terus menaikkan MinArea karena tidak membuang objek kecil secara buta.

## 5. Training hanya setelah label cukup

Jika rule scale stability dan boundary merge belum cukup, barulah kita latih `boundary merge scorer` kecil. Inputnya adalah pasangan region dan feature delta: perbedaan luminance/chroma, boundary strength, overlap, gap, area ratio, serta konsistensi lintas scale. Targetnya hanya `merge` atau `keep separate`.

Model ini tidak akan menjadi semantic model. Ia hanya belajar keputusan geometris yang sempit, sehingga CPU-friendly dan mudah diaudit. Split harus image-level: development untuk fit, calibration untuk threshold, dan blind untuk evaluasi terakhir.

## 6. Retest dan gate

Setelah satu perubahan selesai, jalankan B0, B1-old, dan B1-new pada fixture serta blind 300. Perubahan diterima hanya jika:

1. precision atau F1 naik pada minimal dua domain;
2. recall tidak jatuh melewati batas yang ditentukan;
3. false relation rate tidak naik;
4. runtime dan memory tetap dalam budget;
5. tidak ada regression pada COCO128 development set;
6. semua hasil dapat direproduksi dari commit dan manifest.

## Kesimpulan

Jadi langkah berikutnya bukan “training besar”. Urutannya adalah **annotation → baseline metrics → failure classification → scale-stable/boundary-aware merge → training scorer kecil bila perlu → blind retest**. Dengan urutan ini, setiap klaim HeraVision dapat dibuktikan dan kita tahu apakah peningkatan benar-benar datang dari pemahaman visual atau hanya dari mengurangi jumlah output.
