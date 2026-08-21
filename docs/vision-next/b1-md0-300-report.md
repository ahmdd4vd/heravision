# HeraVision Next — B1-MD0 Blind Multi-Domain Benchmark

## Executive status

Milestone B1-MD0 telah dijalankan pada **300 gambar blind** dengan tiga domain sumber: 143 Imagenette validation images, 143 Imagewoof validation images, dan 14 gambar diagram/dokumen Wikimedia Commons. Semua sample memiliki SHA-256, dimensi yang diverifikasi, domain tag, dan source URL. Tidak ada sample dari manifest ini yang dipakai untuk melatih region filter atau memilih threshold.

> Status: **300/300 completed, 0 errors** untuk B0 dan B1 raw maupun B1 filtered.

Benchmark ini adalah blind **runtime/structure benchmark**, bukan semantic accuracy benchmark penuh. Imagenette/Imagewoof validation images dan Wikimedia diagram set tidak memiliki bounding-box ground truth yang seragam pada manifest ini. Karena itu angka di bawah mengukur completion, region density, relation density, runtime, dan agreement geometry B0–B1; angka tersebut tidak boleh dibaca sebagai object detection precision/recall.

## Dataset dan anti-leakage

| Domain | Jumlah | Sumber | Label status |
|---|---:|---|---|
| Imagenette | 143 | fastai Imagenette repository | Class folder only; no bbox used |
| Imagewoof | 143 | fastai Imagenette repository | Class folder only; no bbox used |
| Wikimedia diagrams/documents | 14 | Per-file Commons URLs in manifest | Unlabeled; license review pending |
| **Total** | **300** |  |  |

Imagenette merupakan subset 10 kelas dari ImageNet yang disediakan melalui repository fastai [1]. Imagewoof memakai keluarga dataset/struktur sumber yang sama sebagai natural-image stress set [1]. Gambar Wikimedia dicatat per-file agar source dan review lisensinya tidak hilang; keberadaan di Wikimedia Commons tidak otomatis berarti semua file memiliki lisensi yang sama.

COCO128 dipakai hanya sebagai development/training source untuk region filter. Blind manifest tidak berisi COCO128. Region filter dibekukan pada bobot dan threshold 0,95 sebelum blind run.

## Command dan konfigurasi

Raw run:

```bash
go run ./cmd/vision-eval \
  -manifest experiments/manifests/blind-md0-300.json \
  -output experiments/runs/blind-md0-300-raw \
  -mode general \
  -max-side 256 \
  -legacy-max-pixels 24000000
```

Filtered run memakai tambahan:

```bash
-region-filter experiments/runs/coco128-verified/region-filter.json \
-region-filter-threshold 0.95
```

Run menggunakan satu CPU worker dan memory budget Go 1,8 GiB. `max-side=256` dipakai karena region proposal pixel-level sebelumnya memicu OOM pada resolusi lebih besar; hal ini dicatat sebagai constraint engine, bukan disembunyikan.

## Completion dan runtime global

| Run | Samples | Completed | Errors | Mean B0 ms | Mean B1 ms | B0 regions | B1 regions |
|---|---:|---:|---:|---:|---:|---:|---:|
| B1 raw | 300 | 300 | 0 | 50,45 | 9,32 | 1.351 | 40.515 |
| B1 filtered | 300 | 300 | 0 | 48,35 | 7,57 | 1.351 | 601 |

Runtime adalah hasil satu lingkungan sandbox, bukan klaim hardware universal. Pengurangan B1 dari 40.515 total region raw menjadi 601 region filtered menunjukkan filter kecil berhasil menekan proposal explosion, tetapi tidak membuktikan region yang tersisa benar secara semantic.

## Per-domain result

### Imagenette

| Mode | Mean B0 regions | Mean B1 regions | Mean B1 edges | Mean B0 ms | Mean B1 ms | Mean B0–B1 coverage | Mean matched IoU |
|---|---:|---:|---:|---:|---:|---:|---:|
| Raw | 2,29 | 165,57 | 3.381,97 | 15,90 | 8,73 | 0,2258 | 0,2944 |
| Filtered | 2,29 | 2,27 | 4,84 | 14,71 | 6,40 | 0,1401 | 0,2293 |

### Imagewoof

| Mode | Mean B0 regions | Mean B1 regions | Mean B1 edges | Mean B0 ms | Mean B1 ms | Mean B0–B1 coverage | Mean matched IoU |
|---|---:|---:|---:|---:|---:|---:|---:|
| Raw | 3,42 | 105,83 | 1.757,53 | 19,06 | 7,45 | 0,1650 | 0,2796 |
| Filtered | 3,42 | 1,52 | 2,10 | 17,96 | 6,29 | 0,1317 | 0,2438 |

### Wikimedia diagrams/documents

| Mode | Mean B0 regions | Mean B1 regions | Mean B1 edges | Mean B0 ms | Mean B1 ms | Mean B0–B1 coverage | Mean matched IoU |
|---|---:|---:|---:|---:|---:|---:|---:|
| Raw | 38,14 | 121,79 | 2.275,43 | 724,07 | 34,43 | 0,0024 | 0,0805 |
| Filtered | 38,14 | 4,29 | 19,36 | 702,43 | 32,64 | 0,0012 | 0,0419 |

Wikimedia diagrams menunjukkan domain yang paling berbeda dari training COCO128. B1 raw menghasilkan banyak edge dan region tetapi agreement dengan B0 rendah; filter menekan jumlah proposal secara agresif. Tanpa annotation diagram yang seragam, kita belum boleh menyebut salah atau benar secara object-level. Namun hasil ini cukup untuk menyimpulkan bahwa pipeline geometry sekarang belum memiliki representasi diagram/layout yang kuat.

## Failure yang ditemukan dan diperbaiki selama benchmark

Run awal menghasilkan dua failure karena dua gambar Wikimedia berukuran lebih dari safety limit 12 megapixel. Adapter awal hanya mengatur max-pixels pada B0, sedangkan B1 masih memakai global decoder limit. Perbaikannya adalah membuat `legacy-max-pixels` eksplisit dan menerapkan pixel budget benchmark ke B0 serta B1 secara scoped, lalu menjalankan ulang benchmark. Hasil final menjadi 300/300 tanpa error.

Koleksi Wikimedia juga sempat memiliki satu duplicate image hash karena dua hasil pencarian flowchart identik. Manifest generator sekarang menolak duplicate hash; satu file diganti dengan image unik dan ID sample memakai domain+hash, bukan nomor urut.

## Interpretasi teknis

Pertama, B1 raw masih mengalami proposal explosion pada natural images: rata-rata lebih dari 100 region per sample, meskipun tidak terjadi error. Ini mengonfirmasi bahwa `MaxRegions=256` adalah guard memory, bukan solusi kualitas.

Kedua, filter yang dilatih dari COCO128 dapat melakukan domain shift reduction dalam arti menekan proposal count pada Imagenette, Imagewoof, dan Wikimedia. Namun penurunan region tidak sama dengan peningkatan pemahaman. Pada Wikimedia, filtered B1 tetap memiliki agreement yang sangat rendah dengan B0, sehingga perlu boundary/diagram-specific evidence dan evaluation annotations.

Ketiga, belum ada semantic ground truth. Dataset blind ini belum dapat menjawab apakah HeraVision mengenali “anjing”, “mobil”, “tabel”, “node”, atau relasi semantik. Ia baru menguji apakah engine dapat memproses gambar lintas domain, menghasilkan graph bounded, dan menjaga runtime/provenance.

## Keputusan milestone

B1-MD0 **lulus sebagai benchmark operasional** karena 300/300 sample selesai tanpa error pada konfigurasi tetap dan artefak per-domain tersimpan. B1-MD0 **belum lulus sebagai semantic quality benchmark** karena tidak memiliki bbox/scene-graph ground truth lintas domain.

Langkah selanjutnya yang disetujui secara teknis adalah membuat annotation layer kecil pada blind set: minimal 30 sample per domain untuk bbox/region review dan 10 sample per domain untuk relation review. Setelah annotation layer tersedia, kita dapat menghitung precision/recall per domain dan menentukan apakah boundary merge scorer atau relation scorer layak dilatih.

## References

[1]: https://github.com/fastai/imagenette "fastai/imagenette official repository"
[2]: https://commons.wikimedia.org/ "Wikimedia Commons"
[3]: https://docs.ultralytics.com/datasets/detect/coco128 "COCO128 dataset documentation used only for development/training"
