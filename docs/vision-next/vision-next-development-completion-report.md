# HeraVision Next — Laporan Konsolidasi Pengembangan

**Branch:** `vision-next`  
**Latest commit:** `9c67e58`  
**Repository:** `ahmdd4vd/heravision`  
**Status main:** tidak diubah

## Ringkasan eksekutif

HeraVision Next sekarang merupakan **engine visual reasoning geometry-first yang berjalan di CPU**, bukan lagi sekadar UI detector lama. Engine dapat membuat canonical luminance/chroma view, menghitung evidence pixel-level, mengusulkan region, menyusun hypothesis generik, membuat visible spatial relations, dan menyimpan provenance pada region, hypothesis, relation, serta graph.

Pengembangan end-to-end yang diminta telah dijalankan. Hasilnya nyata dan terukur, tetapi kesimpulan yang jujur adalah: **branch ini belum menjadi production vision engine umum dan belum dapat memahami semua gambar seperti model vision besar**. Ia sudah layak untuk riset terkontrol dan benchmark internal; belum layak menggantikan `main` atau menerima klaim semantic umum.

## Perubahan yang selesai

| Area | Hasil |
|---|---|
| Accepted-only MD1 evaluation | Manifest 22 sample accepted provisional dibuat dan divalidasi |
| Scale stability | Proposer multi-skala dengan normalized IoU consensus dan boundary-gap guard ditambahkan di belakang `-scale-stable` |
| Boundary evidence | Region menyimpan `scale-consensus`, `scale-region`, `scale_support`, dan `scale_count` |
| Region filter | Filter stable baru dilatih hanya dari COCO development split |
| IoU trainer | Bug perhitungan IoU pada trainer diperbaiki sebelum retraining kedua filter |
| Relation pruning | `BuildPruned` opt-in ditambahkan di belakang `-relation-prune`; containment selalu dipertahankan |
| Graph safety | B0 dan B1 graph divalidasi sebelum sample dianggap sukses |
| Run provenance | Summary menyimpan git SHA, Go version, OS/arch, GOMAXPROCS, GOMEMLIMIT, dan konfigurasi |
| Semantic gate | Label semantic tidak ditambahkan karena belum ada ground truth independen |
| Regression | Unit test, `go vet`, dan race detector lulus |

## Hasil development fixture MD1

Fixture ini berisi 22 sample `accepted-provisional` dari 30 sample awal. Tujuh sample `needs-review` dan satu `accepted-ignore` tidak digunakan pada angka di bawah. Status accepted masih berasal dari same-agent second pass, bukan reviewer manusia independen.

| Konfigurasi | Precision | Recall | F1 | Mean prediction/sample |
|---|---:|---:|---:|---:|
| B0 legacy | 0.0267 | 0.4091 | 0.0501 | 15.32 |
| B1-old + corrected filter | 0.2121 | 0.6364 | 0.3182 | 3.00 |
| B1-stable + stable filter | **0.4194** | 0.5909 | **0.4906** | **1.41** |

Per-domain F1 stable dibanding corrected B1-old meningkat dari 0.4000 menjadi 0.4444 pada Imagenette, dari 0.2857 menjadi 0.3333 pada Imagewoof, dan dari 0.3077 menjadi 0.6154 pada Wikimedia diagrams. Recall aggregate turun 4.55 percentage points, sehingga hasil ini adalah **development ablation yang menjanjikan**, bukan final benchmark.

## Blind retest 300 gambar

Blind MD0 berisi 143 Imagenette, 143 Imagewoof, dan 14 Wikimedia diagrams. Tidak ada ground truth region lengkap, sehingga angka berikut mengukur robustness operasional, jumlah graph, waktu, dan diagnostic overlap—bukan precision/recall final.

| Konfigurasi | Selesai | Error | Total region B1 | Mean B1 ms |
|---|---:|---:|---:|---:|
| B1-old + corrected filter | 300/300 | 0 | 576 | 8.04 |
| B1-stable + stable filter | 300/300 | 0 | 287 | 21.52 |
| B1-stable + stable filter + relation prune | 300/300 | 0 | 287 | 20.37 |

Stable path mengurangi output region sekitar 50.2%, tetapi runtime rata-rata sekitar 2.7 kali lebih lambat karena menghitung tiga skala. Semua 300 sample selesai tanpa error pada constraint `GOMAXPROCS=1`, `GOMEMLIMIT=1800MiB`, dan `max-side=256`.

## Relation pruning

Pada 22 accepted fixture, relation pruning mengurangi total edge dari 256 menjadi 224, atau 12.5%, tanpa mengubah region precision, recall, atau F1. Containment tetap dipertahankan. Karena belum ada relation-labeled benchmark, perubahan ini hanya boleh disebut **graph-size optimization**, bukan peningkatan pemahaman relasi.

## Kualitas kode dan reproducibility

Pemeriksaan berikut lulus:

| Pemeriksaan | Hasil |
|---|---|
| `go test ./internal/visionnext/... ./cmd/vision-eval` | Lulus |
| `go vet ./internal/visionnext/... ./cmd/vision-eval` | Lulus |
| `CGO_ENABLED=1 go test -race ./internal/visionnext/...` | Lulus |
| Accepted-only final smoke | 22/22, zero errors |
| Run metadata | Git SHA dan runtime config tersimpan |

Final smoke menggunakan git SHA `77bc2794c4071829aba4e81adc516a9b4b8f2578` sebelum dokumentasi commit terakhir. Semua source commit telah diverifikasi pada branch; run artifacts tetap tidak dimasukkan ke Git.

## Status kesiapan

| Kategori | Status |
|---|---|
| Prototype research engine | Siap |
| CPU-controlled internal benchmark | Siap |
| General image understanding | Belum |
| Semantic object/activity recognition | Belum |
| Production API/service | Belum |
| Final public benchmark claim | Belum |
| Penggantian `main` | Tidak disarankan |

> Kesimpulan sederhana: HeraVision Next sekarang sudah menjadi **mesin riset visual yang sungguh-sungguh berjalan dan dapat diaudit**, tetapi belum menjadi “mata AI” yang memahami semua gambar.

## Blocker sebelum production

Pertama, tujuh disagreement MD1 harus diputuskan oleh reviewer independen. Kedua, perlu benchmark berlabel yang lebih besar dan mencakup region, relation, abstention, serta semantic label bila semantic layer akan ditambahkan. Ketiga, perlu API yang stabil, load test, memory test pada berbagai ukuran gambar, timeout, dan dokumentasi supported input. Keempat, stable path harus dibuktikan tidak mengorbankan recall terlalu besar pada benchmark independen.

Semantic prior sengaja belum ditambahkan. Tanpa label independen dan evidence yang cukup, menamai region sebagai `dog`, `person`, atau `car` akan menjadi klaim yang tidak dapat diaudit dan bertentangan dengan kontrak evidence-first.

## Commit utama

| Commit | Isi |
|---|---|
| `3637cd4` | Accepted-only MD1 baseline evaluation |
| `e846b0a` | Experimental scale-stable proposals |
| `1d7b740` | Opt-in relation spatial pruning |
| `1042a7c` | Corrected IoU trainer and stable filter report |
| `8f84c05` | Stable blind retest report |
| `77bc279` | Graph validation and run metadata |
| `9c67e58` | Production readiness report |

## Artefak penting

- `docs/vision-next/md1-accepted-22-baseline-report.md`
- `docs/vision-next/b1-scale-stability-ablation.md`
- `docs/vision-next/relation-spatial-pruning-ablation.md`
- `docs/vision-next/region-filter-stable-training.md`
- `docs/vision-next/b1-md0-300-stable-retest-report.md`
- `docs/vision-next/semantic-gate-decision.md`
- `docs/vision-next/production-readiness-report.md`
- `experiments/manifests/md1-ground-truth-eval-accepted-22.json`
- `experiments/compare_blind_runs.py`
- `experiments/analyze_relation_runs.py`
