# HeraVision Next — Laporan Fase Lanjutan

**Branch:** `vision-next`  
**Latest commit:** `611b5e2`  
**Main branch:** unchanged

## Ringkasan

Fase lanjutan memperkuat HeraVision Next tanpa menambahkan semantic label yang belum terbukti. Perubahan utama adalah answer evidence-first, fixture abstention, fixture review relation, eksperimen scale support, eksperimen extra scale, serta hardening decode agar batas pixel aman untuk concurrent request.

## Answer dan abstention

Output B0/B1 sekarang menyertakan `Answer` dengan tiga status: `answered`, `abstain`, dan `insufficient_evidence`. Teks yang diizinkan hanya klaim generic `stable visual structure detected`; sistem tidak menyebut nama objek. Answer menyalin provenance region sehingga klaim dapat ditelusuri kembali.

| Run | Samples | Errors | Answered | Abstain | Insufficient evidence |
|---|---:|---:|---:|---:|---:|
| Accepted MD1 stable | 22 | 0 | 17 | 0 | 5 |
| Blind MD0 stable | 300 | 0 | 233 | 0 | 67 |
| MD1 candidate fixture | 30 | 0 | 24 | 0 | 6 |

Tidak adanya status `abstain` menunjukkan threshold lemah belum terpicu pada data ini; fixture answerability independen masih diperlukan.

## Relation review

Fixture relation independen telah dibuat dari 30 sample MD1 dan berisi 50 predicted edges dengan endpoint bbox dan predicate. Reviewer dapat menandai `correct`, `incorrect`, atau `uncertain`. Fixture ini mengukur false relation rate/predicted-edge precision, bukan relation recall karena missed edges belum dianotasi.

## Scale ablations

Minimum support satu skala diuji dengan filter baru dari COCO. Pada MD1, support satu skala menghasilkan F1 0.2600, lebih rendah daripada stable default support dua skala dengan F1 0.4906. Extra scale 1.25x juga diuji; F1 menjadi 0.3095 dan runtime naik dari 30.59 ms menjadi 38.55 ms. Kedua perubahan tidak dipromosikan sebagai default.

| Variant | Precision | Recall | F1 | Mean B1 ms |
|---|---:|---:|---:|---:|
| Stable default, support >=2 | 0.4194 | 0.5909 | 0.4906 | 30.59 |
| Stable support >=1 | 0.1667 | 0.5909 | 0.2600 | 31.27 |
| Stable + extra scale 1.25 | 0.2097 | 0.5909 | 0.3095 | 38.55 |

## Robustness

Pada real MD1 image set, max-side 64, 256, dan 512 semuanya selesai tanpa error. Namun max-side 64 menghasilkan nol region dari 30 sample, sehingga aman secara operasional tetapi tidak layak sebagai konfigurasi kualitas. Max-side 256 tetap rekomendasi utama; max-side 512 lebih lambat dan memerlukan benchmark kualitas tersendiri.

Manifest valid-plus-corrupt juga diuji. Sample valid berhasil diproses, sedangkan file rusak ditandai `status=error` tanpa menghentikan keseluruhan run.

## Concurrency hardening

Global mutable `processor.MaxPixels` tidak lagi dimodifikasi oleh adapter B0/B1. Ditambahkan `DecodeWithMaxPixels` untuk limit per request, dan `facts.Extract` serta B1 memakai limit tersebut. Full tests, `go vet`, dan race detector pada processor, facts, serta visionnext lulus.

## Keputusan semantic dan production

Semantic layer tetap **belum diaktifkan**. Tujuh disagreement MD1 belum memiliki keputusan reviewer independen, relation ground truth belum selesai, answerability fixture belum direview independen, dan blind 300 belum menyediakan precision/recall lengkap. Karena itu branch ini tetap research-ready, bukan production-ready.

> Kemajuan fase ini membuat HeraVision lebih jujur dan lebih aman: ia dapat menjelaskan kapan evidence ada, kapan evidence tidak cukup, dan kapan input gagal. Ia belum menjadi mesin semantic yang memahami semua gambar.

## Artefak utama

- `internal/visionnext/answer/answer.go`
- `experiments/analyze_abstention.py`
- `experiments/manifests/md1-abstention-fixture-30.json`
- `experiments/manifests/md1-relation-review-fixture.json`
- `docs/vision-next/scale-support-ablation-report.md`
- `docs/vision-next/extra-scale-ablation-report.md`
- `docs/vision-next/robustness-concurrency-report.md`
- `experiments/analyze_robustness_sweep.py`
- `internal/processor/decode.go`

## Dependency yang masih menunggu user/reviewer

Untuk melanjutkan ke benchmark final, reviewer independen perlu mengisi `experiments/manifests/md1-independent-review-template.json`, `experiments/manifests/md1-abstention-fixture-30.json`, dan `experiments/manifests/md1-relation-review-fixture.json`. Tanpa itu, saya akan terus menjalankan engineering ablations, tetapi tidak akan menaikkan status metrik menjadi klaim final.
