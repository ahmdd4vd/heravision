# MD1 Ground-Truth Region — Provisional Report

## Status penting

Ground truth region untuk 30 fixture MD1 sudah dibuat dan divalidasi secara programmatic, tetapi statusnya **provisional analyst review**, bukan final independent human ground truth. Box dibuat dari review visual contact sheet, kemudian diverifikasi terhadap hash/dimensi image dan batas canvas. Sebelum dipakai sebagai klaim publik, subset ini tetap memerlukan review kedua.

## Coverage

| Item | Jumlah |
|---|---:|
| Fixture | 30 |
| Fixture dengan region | 29 |
| Fixture `ignore` | 1 |
| Total region provisional | 29 |
| Region ID unique | 29 |
| Hash mismatch | 0 |
| Box out-of-bounds | 0 |

Blank scan diberi `ignore_reason=no visible structure`. Region lain memakai `whole_diagram`, `main_object`, `main_object_partial`, atau `scene_structure`. Tidak ada class folder Imagenette/Imagewoof yang dipakai sebagai semantic ground truth.

## Provisional fixture metrics

Predictions berasal dari B1-new MinArea=8 pada 30 fixture yang sama. Matching menggunakan one-to-one greedy IoU dengan threshold 0,50.

| System | Pred count | TP | FP | FN | Precision | Recall | F1 | Mean IoU matched |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| B0 legacy | 374 | 11 | 363 | 18 | 0,0294 | 0,3793 | 0,0546 | 0,7605 |
| B1 raw | 1.920 | 20 | 1.900 | 9 | 0,0104 | 0,6897 | 0,0205 | 0,7152 |
| B1 filtered | 92 | 18 | 74 | 11 | **0,1957** | 0,6207 | **0,2975** | 0,7186 |

Angka ini hanya diagnosis awal karena anotasi belum independent. Meski demikian, pola yang terlihat konsisten dengan MD0: B1 raw memiliki recall lebih tinggi tetapi proposal explosion, sedangkan B1 filtered mengurangi false positives dan meningkatkan F1 provisional dibandingkan B0.

## Aturan anotasi

Region harus memiliki batas visible yang dapat dijelaskan. Jika batas tidak cukup jelas, reviewer boleh memberi visibility `uncertain`; jika tidak ada struktur yang layak dinilai, fixture memakai `ignore`. Diagram dibedakan antara whole-canvas region dan subregion hanya ketika batas internalnya jelas. Relation semantic belum dianotasi pada tahap ini.

## Validasi yang sudah dijalankan

Generator memverifikasi SHA-256 setiap image terhadap parent manifest. Validator memastikan 30 sample unique, 29 region ID unique, semua box memiliki ukuran positif, seluruh box berada dalam dimensi image, dan setiap region membawa `source=analyst-visual-review` serta `status=provisional-review`.

## Gate berikutnya

Sebelum memakai angka ini sebagai benchmark final, lakukan independent second review pada 30 fixture, selesaikan disagreement log, dan ubah status hanya untuk box yang disetujui. Setelah itu metrik dapat dipakai untuk memilih antara boundary-aware merge, scale stability, atau boundary merge scorer kecil.
