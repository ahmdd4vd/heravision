# HeraVision Next — B1-MD1 MinArea Ablation

## Scope

B1-MD1 menguji satu perubahan terisolasi pada failure family **over-segmentation**: default `MinArea` region proposal dinaikkan dari 4 menjadi 8 pada canonical view 256px. Tidak ada perubahan pada merge threshold, filter weights, threshold 0,95, manifest, dataset, atau runtime command.

Blind manifest tetap berisi 300 sample yang sama dan tidak dipakai untuk training atau threshold tuning. B1-old merujuk run sebelum perubahan; B1-new merujuk run dengan `MinArea=8`.

## Global result

| Variant | Samples | Errors | B1 raw regions | B1 filtered regions | Mean B1 raw ms |
|---|---:|---:|---:|---:|---:|
| B1-old | 300 | 0 | 40.515 | 601 | 9,32 |
| B1-new MinArea=8 | 300 | 0 | 22.270 | 601 | 7,90 |

`MinArea=8` mengurangi raw region proposal sekitar 45,0% dan menurunkan raw B1 runtime. Jumlah filtered region tetap 601 karena model filter dan threshold yang sama memilih subset akhir yang sama pada blind run ini.

## Per-domain raw effect

| Domain | B1-old regions/sample | B1-new regions/sample | Reduction |
|---|---:|---:|---:|
| Imagenette | 165,57 | 96,78 | 41,5% |
| Imagewoof | 105,83 | 53,08 | 49,8% |
| Wikimedia diagrams | 121,79 | 60,00 | 50,7% |

Edges ikut berkurang dari 3.381,97 menjadi 1.891,78 pada Imagenette, dari 1.757,53 menjadi 839,28 pada Imagewoof, dan dari 2.275,43 menjadi 1.005,64 pada Wikimedia diagrams. Ini menunjukkan perubahan mempengaruhi fragmentasi graph, bukan hanya serialization.

## Quality interpretation

Perubahan ini adalah **structural improvement**, bukan bukti semantic improvement. Pada filtered run, coverage dan mean matched IoU tetap sama karena final retained regions tidak berubah. Karena blind set belum memiliki bbox/scene-graph ground truth, kita tidak boleh menyimpulkan recall atau precision meningkat.

Perubahan diterima sebagai guard kualitas awal karena:

1. completion tetap 300/300 dengan 0 error;
2. raw proposal count turun sekitar 45%;
3. runtime raw B1 turun;
4. filtered output tidak memburuk pada metrik geometry agreement yang tersedia;
5. unit dan integration tests tetap lulus.

Namun perubahan belum menyelesaikan masalah utama: B1 masih tidak memiliki semantic ground truth lintas domain dan relation graph masih berpotensi terlalu padat. Annotation fixture MD1 tetap diperlukan.

## Keputusan engineering

`MinArea=8` dipertahankan sebagai default sementara pada branch `vision-next`, dengan catatan bahwa nilai ini bukan hyperparameter final. Nilai ideal harus dikalibrasi per resolution dan domain setelah 90 region-reviewed annotations tersedia.

Eksperimen berikutnya sebaiknya menargetkan **boundary-aware merge atau scale stability**, bukan menaikkan `MinArea` terus-menerus. Menaikkan area tanpa evidence dapat menghapus object kecil dan menurunkan recall secara diam-diam.
