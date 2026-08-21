# HeraVision Next — COCO128 Pilot Result

## Status

Eksperimen ini adalah **pilot terverifikasi**, bukan benchmark general vision final. Dataset berasal dari asset COCO128 resmi Ultralytics. Arsip berisi 128 image JPEG dan 128 file label, tetapi dua image stem dan dua label stem tidak berpasangan. Untuk mencegah ground-truth salah, hanya 126 pasangan yang digunakan; dua image tanpa label dan dua label tanpa image dicatat sebagai exclusion di manifest.

Manifest dan protocol menyimpan source URL, archive SHA-256, image SHA-256, dimensi, dan ground-truth boxes. Dataset binary tidak dimasukkan ke repository.

## Pipeline yang diuji

| Sistem | Deskripsi |
|---|---|
| B0 | Engine legacy `facts.Extract`, default legacy config |
| B1 raw | Canonical view + evidence field + region proposal, tanpa model filter |
| B1 filtered | B1 raw + logistic region filter 13 fitur, dilatih pada 80% image IDs dan diuji pada 20% image IDs |

B1 dijalankan dengan `max-side=256`, `GOMAXPROCS=1`, dan `GOMEMLIMIT=1800MiB` untuk menjaga eksperimen CPU/RAM-friendly. B0 dan B1 memakai manifest image yang sama. Ground-truth comparison menskalakan box output ke ukuran image asli dan memakai one-to-one greedy IoU matching.

## Raw batch result

Run command:

```bash
go run ./cmd/vision-eval \
  -manifest experiments/manifests/coco128-verified.json \
  -output experiments/runs/coco128-verified \
  -mode general -max-side 256
```

Hasil batch: **126/126 completed, 0 errors**. B0 menghasilkan 2.321 region, sedangkan B1 raw menghasilkan 25.562 region. Runtime mean B0 adalah 201,34 ms/sample dan B1 raw 14,25 ms/sample pada run tersebut. Angka runtime bersifat lingkungan-dependent dan belum menjadi benchmark hardware final.

### Ground-truth IoU = 0,50

| System | Pred boxes | TP | FP | FN | Precision | Recall | F1 | Mean IoU matched |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| B0 | 2.321 | 70 | 2.251 | 859 | 0,0302 | 0,0753 | 0,0431 | 0,7044 |
| B1 raw | 25.562 | 124 | 25.438 | 805 | 0,0049 | 0,1335 | 0,0094 | 0,6844 |
| B1 filtered | 464 | 70 | 394 | 859 | 0,1509 | 0,0753 | 0,1005 | 0,7096 |

Interpretasi yang benar adalah B1 raw memiliki recall lebih tinggi tetapi fragmentasi/false-positive sangat buruk. Filter terlatih mengurangi output rata-rata dari 202,87 menjadi 3,68 region/sample dan meningkatkan precision serta F1 dibandingkan B0 pada IoU 0,50, tetapi recall belum meningkat. Ini adalah peningkatan **quality of proposals**, bukan bukti pemahaman semantic.

## Sensitivity terhadap IoU threshold

| IoU threshold | B0 F1 | B1 raw F1 | B1 filtered F1 |
|---:|---:|---:|---:|
| 0,25 | 0,0837 | 0,0228 | 0,1823 |
| 0,50 | 0,0431 | 0,0094 | 0,1005 |
| 0,75 | 0,0185 | 0,0029 | 0,0416 |

Pada ketiga threshold, B1 filtered lebih baik daripada B1 raw dan B0 dalam F1 pilot ini. Namun hasil ini harus dianggap sebagai hasil subset COCO128, bukan generalisasi lintas foto, dokumen, diagram, screenshot, ilustrasi, atau gambar ambigu.

## Training detail

Model yang dilatih bukan VLM dan bukan object classifier. Targetnya adalah binary region quality filter: sebuah proposal positif bila IoU terbaiknya terhadap ground-truth box mencapai threshold 0,50. Fitur yang dipakai hanya 13 feature geometry/evidence: log area, log width, log height, aspect ratio, area ratio, compactness, boundary strength, scale stability, posisi, ukuran relatif, dan texture.

Split dilakukan pada image ID sebelum training sehingga region dari satu image tidak tersebar ke train dan test. Logistic regression memakai standardization, positive-class weighting, 2.500 gradient steps, dan threshold dipilih dari train split saja. Threshold yang terpilih adalah 0,95. Test region berisi 20% image ID yang tidak dipakai saat fit.

Test filter terlatih menghasilkan precision 0,2075, recall 0,6111, F1 0,3099 pada level region filter. Angka ini bukan sama dengan final detection F1 karena setelah filter masih dilakukan matching per image terhadap semua ground-truth boxes.

## Failure dan keputusan engineering

Batch pertama pada `max-side=512` mengalami termination karena region proposal pixel-level dan relation builder menghasilkan terlalu banyak komponen/pasangan. Ini adalah failure nyata, bukan disembunyikan. Perbaikan yang diterapkan adalah `MinArea=4` dan `MaxRegions=256` pada region proposal, lalu pilot berhasil 126/126 tanpa error pada `max-side=256`. Perbaikan tersebut harus diuji lagi pada resolusi lebih tinggi sebelum dianggap selesai.

B1 raw menghasilkan banyak proposal kecil. Ini menunjukkan bottleneck utama saat ini adalah **region formation dan proposal calibration**, bukan semantic understanding. Karena itu training berikutnya sebaiknya tidak langsung berupa captioner. Prioritasnya adalah boundary-aware merge, scale stability, spatial pruning untuk relation candidate, dan evaluasi lintas domain.

## Kesimpulan yang dapat dipertahankan

Pada pilot terverifikasi ini, B1 raw belum lebih baik secara keseluruhan daripada B0; ia memiliki recall lebih tinggi tetapi precision/F1 lebih buruk karena over-segmentation. Setelah komponen filter kecil dilatih dengan split image-level, B1 filtered mengalahkan B0 dalam F1 pada IoU 0,25, 0,50, dan 0,75 pada subset ini. Hasil tersebut valid sebagai bukti bahwa **training komponen kecil dapat memperbaiki proposal quality**, tetapi belum membuktikan HeraVision memahami gambar secara umum.

## Langkah berikutnya

1. Memasukkan manifest dan evaluator ke regression suite tanpa memasukkan dataset binary.
2. Menambah domain dokumen, diagram, screenshot, dan low-quality image sehingga pilot tidak hanya foto COCO.
3. Menjalankan blind test yang tidak dipakai untuk threshold tuning.
4. Memperbaiki region merge dan relation candidate pruning berdasarkan failure gallery.
5. Baru setelah itu membandingkan semantic prior compact atau training relation scorer kecil.
