# Setelah B1-MD0: Langkah Berikutnya

## Kesimpulan posisi saat ini

B1-MD0 sudah membuktikan bahwa pipeline dapat memproses 300 gambar lintas sumber tanpa error dan bahwa region filter mampu mengurangi proposal explosion. Namun benchmark tersebut belum memiliki anotasi region/relation yang seragam, sehingga kita belum tahu apakah region yang dipertahankan benar-benar benar. Tahap berikutnya adalah mengukur **kualitas**, bukan menambah jumlah fitur.

## Tahap 1 — Annotation layer terukur

Pilih subset terstratifikasi dari blind MD0:

| Domain | Region review | Relation review |
|---|---:|---:|
| Imagenette | 30 | 10 |
| Imagewoof | 30 | 10 |
| Wikimedia diagram/document | 30 | 10 |
| **Total** | **90** | **30** |

Untuk setiap gambar region review, anotasi minimum adalah bounding region yang terlihat, bukan label semantic yang dipaksakan. Untuk relation review, gunakan predicate visible-only seperti `above`, `left_of`, `inside`, `overlapping`, dan `touching`. Pisahkan anotasi development dan blind evaluation; jangan mengubah threshold berdasarkan blind evaluation.

## Tahap 2 — Failure gallery

Dari output B1 raw dan filtered, kelompokkan kegagalan menjadi fragmentation, merge error, missed region, boundary leak, false relation, low-quality collapse, atau runtime/memory. Buat contact sheet yang menampilkan input, B0, B1 raw, B1 filtered, dan anotasi. Setiap perbaikan harus memilih satu failure family, bukan mengubah banyak komponen sekaligus.

## Tahap 3 — Perbaikan algoritma paling penting

Urutan teknis yang paling masuk akal adalah:

1. Mengganti `MaxRegions` sebagai pemotong kasar dengan region budget berbasis information value.
2. Menambahkan scale stability dan boundary-aware merge agar region kecil yang tidak stabil tidak memenuhi graph.
3. Mengganti relation all-pairs dengan spatial candidate pruning.
4. Menyimpan polygon atau evidence boundary agar relation tidak hanya bergantung pada bounding box.
5. Mengukur ulang B1-old versus B1-new pada development set.

Jika perubahan tidak meningkatkan metrik pada annotation layer, perubahan harus di-revert meskipun output terlihat lebih kompleks.

## Tahap 4 — Training berikutnya hanya jika dibenarkan

Filter region COCO128 sudah dilatih dan dibekukan. Model berikutnya tidak boleh dilatih sebelum failure gallery menunjukkan kebutuhan yang jelas. Kandidat pertama adalah boundary merge scorer kecil atau relation scorer kecil, bukan captioner/VLM besar. Training harus menggunakan development annotations, calibration split untuk threshold, dan blind subset yang tidak pernah dipakai tuning.

## Tahap 5 — Blind retest

Setelah satu perubahan terukur selesai, jalankan ulang 300 MD0 dengan commit baru dan konfigurasi yang sama. Bandingkan completion, region count, relation count, runtime, precision/recall pada 90 annotated samples, dan unsupported relation rate. Hasil yang hanya membaik pada COCO128 tetapi memburuk pada Wikimedia tidak boleh dianggap sebagai kemenangan global.

## Gate sebelum model semantic

Semantic prior compact baru boleh masuk jika tiga kondisi terpenuhi:

| Gate | Syarat |
|---|---|
| Evidence | Region dan relation visible sudah stabil pada tiga domain |
| Data | Ada annotation semantic/legal yang cukup untuk task terbatas |
| CPU | Model memenuhi budget latency/memory dan tidak menaikkan unsupported claim rate |

Sampai gate tersebut terpenuhi, fokus HeraVision tetap evidence-first geometry dan provenance.

## Milestone berikutnya

Target berikutnya adalah **B1-MD1: annotated quality benchmark**, bukan “AI memahami semua gambar”. Milestone ini selesai jika 90 region-reviewed samples dan 30 relation-reviewed samples memiliki anotasi konsisten, B0/B1 memiliki precision/recall per domain, satu failure family diperbaiki dengan ablation, dan blind 300 retest dapat direproduksi dari commit serta manifest yang sama.
