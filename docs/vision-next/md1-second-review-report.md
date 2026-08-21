# MD1 Second Review and Disagreement Log

## Review status

Second review terhadap 30 fixture telah selesai sebagai **same-agent second pass**. Anotasi provisional awal tetap dipertahankan; second review hanya menambahkan status dan alasan. Karena reviewer kedua masih Manus AI pada sesi yang sama, hasil ini belum dapat disebut independent human agreement.

## Summary

| Status | Samples |
|---|---:|
| Accepted provisional | 22 |
| Needs review | 7 |
| Accepted ignore | 1 |
| **Total** | **30** |

Satu sample blank scan dipertahankan sebagai `accepted-ignore` karena tidak memiliki struktur visible yang layak dianotasi. Tujuh sample masuk `needs-review` dan tidak boleh dipakai sebagai final benchmark tanpa keputusan reviewer independen.

## Disagreement categories

| Category | Samples | Reason |
|---|---:|---|
| `scope-ambiguity` | 2 | Box dapat dibaca sebagai scene/object scope yang berbeda |
| `multiple-foreground` | 2 | Ada lebih dari satu foreground yang masuk akal |
| `occlusion-scope` | 2 | Person/object overlap atau crop membuat ownership region tidak jelas |
| `diagram-scope` | 1 | Whole-diagram box mencakup whitespace/connector extent yang perlu disepakati |

Sample IDs dan alasan lengkap tersedia di `experiments/manifests/md1-disagreement-log.json`. Tidak ada disagreement yang menghapus atau menimpa anotasi provisional; setiap perubahan dapat dilacak melalui `second_review` dan `second_review_category`.

## Accepted criteria

Box diterima secara provisional bila batas region terlihat konsisten dengan `region_type`, tidak memerlukan interpretasi semantic, dan tidak bersaing dengan foreground lain yang sama kuat. Accepted di sini berarti **accepted by this second pass**, bukan gold-standard final.

Sample ambigu tetap dipisahkan agar precision/recall tidak tercemar oleh keputusan arbitrer. Setelah reviewer independen memutuskan, sample `needs-review` dapat dipindahkan ke accepted, rejected, atau ignore dengan alasan baru dan timestamp review.

## Gate untuk final metrics

Final benchmark MD1 belum dibuka. Sebelum memakai subset ini sebagai klaim kualitas, diperlukan second reviewer independen atau user review untuk tujuh disagreement. Bila tidak ada reviewer kedua, subset accepted 22 dapat dipakai sebagai development diagnostic dengan label `provisional`, sedangkan tujuh sample tetap excluded dari headline metrics.

## Kesimpulan

Second pass menemukan disagreement yang bermakna, bukan sekadar mengonfirmasi semua box. Hal ini menunjukkan schema `needs-review` bekerja dan mencegah anotasi ambigu diperlakukan sebagai ground truth. Langkah berikutnya adalah menyelesaikan tujuh kasus tersebut melalui review independen, lalu menghitung ulang metrik pada accepted subset dan melakukan disagreement audit.
