# Knowledge-Compressed Speculative Vision for HeraVision

## Jawaban inti

**Ada celah optimasi yang nyata**, tetapi celah tersebut bukan formula yang menghapus kebutuhan akan pengetahuan. Celahnya adalah mengatur **kapan pengetahuan dipakai, seberapa banyak yang dipakai, dan bagaimana jawaban murah diverifikasi**.

Speculative decoding bekerja karena model kecil membuat draft, lalu model target memverifikasi draft. Jika benar, pekerjaan murah dipakai; jika salah, draft dibuang. Google menjelaskan bahwa teknik ini dapat mempertahankan distribusi output target sambil menambah paralelisme [1]. DeepSeek DSpark memperluas gagasan tersebut dengan draft semi-autoregressive dan verification length yang dijadwalkan berdasarkan confidence serta throughput [4].

Untuk HeraVision, analoginya adalah:

```text
cheap visual proposal
→ knowledge retrieval kecil
→ evidence verification
→ accept / refine / abstain
```

Namun ada perbedaan fundamental. LLM sudah memiliki target model yang dapat memverifikasi token. HeraVision tidak otomatis memiliki “model target” yang mengetahui semua benda. Karena itu, verification harus berasal dari kombinasi **dataset berlabel, visual evidence, relation consistency, dan knowledge graph**, bukan dari tebakan kedua yang sama-sama tidak berpengetahuan.

## 1. Tiga hal yang harus dipisahkan

Masalah “ini hewan apa?” sebenarnya terdiri atas tiga masalah berbeda.

| Lapisan | Pertanyaan | Sumber jawaban |
|---|---|---|
| Visual evidence | Region mana yang terlihat? | Pixel, boundary, texture, geometry |
| Semantic recognition | Kategori apa yang paling cocok? | Classifier/embedding terlatih |
| World knowledge | Apa arti kategori itu dan relasinya? | Knowledge graph/retrieval |

B1 saat ini kuat terutama pada lapisan pertama. Knowledge graph dapat memberi fakta bahwa `dog` adalah `animal`, tetapi graph tidak dapat menentukan bahwa region pada foto adalah dog tanpa recognizer visual. Sebaliknya, recognizer yang berkata `dog` tanpa region dan provenance juga tidak cukup evidence-first.

Jawaban aman harus berbentuk:

```text
visual evidence → candidate label → knowledge constraints → verified answer
```

## 2. Jangan menyimpan semua pengetahuan di neural model

Untuk PC kentang, pengetahuan dunia sebaiknya dibagi menjadi tiga bentuk.

### A. Ontology kecil dan stabil

Ontology berisi struktur kategori, misalnya:

```text
entity
├── living_thing
│   ├── animal
│   │   ├── mammal
│   │   │   ├── dog
│   │   │   └── cat
│   │   └── bird
│   └── plant
├── person
├── vehicle
└── artifact
```

Ini murah disimpan sebagai adjacency list atau integer IDs. Ontology tidak membuat model bisa mengenali gambar; ia hanya mencegah jawaban semantic saling bertentangan.

### B. Prototype dan feature signature

Setiap kategori memiliki prototype feature ringkas:

```text
prototype(dog) = [shape, aspect, texture, color, pose, context]
```

Prototype bukan foto mentah dan bukan caption panjang. Ia dapat berupa centroid dan covariance dari embedding/feature vector. Untuk kategori `k`, simpan:

```text
μ_k = mean feature vector
Σ_k = covariance atau diagonal variance
prior_k = frekuensi development
```

Jarak Mahalanobis memberi skor kecocokan murah:

```text
d_k(z) = (z - μ_k)^T Σ_k^-1 (z - μ_k)
```

Untuk CPU, Σ dapat dibatasi diagonal sehingga biaya menjadi linear terhadap jumlah fitur:

```text
d_k(z) = Σ_j (z_j - μ_kj)^2 / (σ_kj^2 + ε)
```

### C. Retrieval facts yang hanya dipanggil jika perlu

Fakta seperti `dog is an animal` atau `vehicle can have wheels` disimpan sebagai triples terkompresi:

```text
(subject_id, relation_id, object_id)
```

Retriever tidak mencari seluruh graph. Ia hanya mengambil subgraph yang terkait dengan kandidat top-k. Jika visual classifier menghasilkan `dog`, engine mengambil ancestor `animal`, sibling dekat `cat`, dan constraint relevan. Ini jauh lebih ringan daripada memasukkan seluruh knowledge base ke setiap inferensi.

## 3. Mathematical cascade

Misalkan `z` adalah feature vector region dari B1. Classifier broad menghasilkan skor kandidat:

```text
s_k = w_k · z + b_k
p_k = sigmoid(s_k)
```

Skor ini belum cukup. Kita kalikan dengan quality factors:

```text
E_k = p_k × q_region × q_boundary × q_scale × q_input
```

Kemudian gunakan margin:

```text
m = E_top1 - E_top2
```

Keputusan:

```text
if no_valid_region:
    insufficient_evidence
else if E_top1 < τ or m < δ:
    abstain
else:
    answered(top1)
```

`τ` mengontrol keamanan absolut. `δ` mencegah jawaban ketika `dog` dan `cat` sama-sama mungkin.

Untuk beberapa region, evidence dapat digabung dengan noisy-OR:

```text
support_k = 1 - Π_i (1 - E_k,i)
```

Region yang tumpang tindih berat perlu diberi redundancy penalty agar lima pecahan dari objek yang sama tidak dihitung sebagai lima bukti independen.

## 4. Gagasan “speculative vision” yang dapat diuji

Kita dapat membuat tiga jenis proposal.

| Proposal | Biaya | Contoh |
|---|---:|---|
| P0 | Sangat murah | warna, gradient, aspect ratio, region geometry |
| P1 | Sedang | texture, local pattern, prototype distance |
| P2 | Mahal | OCR, face specialist, crop classifier, relation verification |

P0 membuat hipotesis kasar. P1 memeriksa apakah hipotesis itu masuk akal. P2 hanya dijalankan bila P0/P1 memberi signal yang cukup.

Contoh:

```text
P0 melihat region compact dengan texture berbulu
→ kandidat {animal, plush_object}
→ P1 mengambil prototype dan context
→ jika margin cukup, jawab broad animal
→ jika dog vs cat belum jelas, ambil P2 atau abstain
```

Ini mirip draft/verify, tetapi tidak boleh disebut lossless speculative decoding. Di vision, proposal yang salah dapat dibuang, tetapi verification tidak menjamin label benar jika target knowledge belum dilatih.

## 5. Confidence-scheduled verification

DSpark menjadwalkan panjang verification berdasarkan survival probability dan throughput profile [4]. HeraVision dapat memakai ide analog:

```text
p_survive = P(proposal tetap benar setelah verifier)
```

Jika `p_survive` tinggi, cukup verifikasi boundary dan relation utama. Jika rendah, periksa crop tambahan, scale lain, OCR, atau specialist. Jika biaya verifikasi lebih besar daripada nilai informasi, abstain.

Secara sederhana, pilih aksi `a` yang memaksimalkan expected utility per biaya:

```text
a* = argmax_a [ ΔRiskReduction(a) / Cost(a) ]
```

`ΔRiskReduction` adalah perkiraan penurunan risiko jawaban salah; `Cost` adalah waktu CPU, memory, dan I/O. Ini membuat engine tidak selalu memakai pipeline paling mahal.

## 6. Early exit dan information gain

Kita dapat berhenti lebih awal jika:

```text
E_top1 >= τ_high
m >= δ_high
provenance_complete = true
```

Sebaliknya, jika skor tidak berubah setelah pemeriksaan tambahan, tidak perlu memanggil specialist lagi. Untuk memilih crop berikutnya, gunakan estimasi information gain:

```text
IG(a) = H(current_belief) - E[H(belief_after_a)]
```

Pilih crop atau scale yang diperkirakan paling banyak mengurangi entropy. Pada implementasi awal, kita tidak perlu menghitung entropy penuh; proxy sederhana adalah margin label dan ketidakstabilan region.

## 7. Representasi pengetahuan yang ringan

Jangan membuat satu model raksasa yang memuat semua fakta. Gunakan:

| Komponen | Representasi CPU |
|---|---|
| Ontology | integer IDs + adjacency list |
| Facts | compressed triples |
| Category prototypes | mean + diagonal variance |
| Visual rules | small logistic/tree scorer |
| Provenance | region IDs + evidence refs |
| Hard negatives | hashes dan sample IDs |
| Specialist routing | decision table/threshold |

Dengan desain ini, pengetahuan yang sering dipakai berada di cache. Pengetahuan langka diambil hanya ketika kandidat memerlukannya. Biaya inference menjadi lebih dekat ke:

```text
Cost ≈ Cost(P0) + P(P1)·Cost(P1) + P(P2)·Cost(P2)
```

Jika hanya 10% gambar membutuhkan P2, rata-rata biaya jauh lebih rendah daripada menjalankan P2 pada 100% gambar.

## 8. “Semua hewan” tetap membutuhkan data

Ontology dapat mencakup ribuan hewan dengan biaya memori kecil, tetapi **cakupan nama bukan cakupan kemampuan visual**. Agar model dapat membedakan `dog`, `wolf`, dan `fox`, ia memerlukan prototype atau classifier yang dilatih pada contoh visual yang relevan. Knowledge graph hanya mengetahui hubungan antar nama.

Solusi semi-open-world:

1. Klasifikasikan ke broad category terlebih dahulu.
2. Ambil kandidat top-k dari ontology/prototype bank.
3. Cocokkan feature region dengan prototype kandidat.
4. Gunakan hard-negative comparison.
5. Jawab spesies detail hanya jika margin dan evidence mencukupi.
6. Jika tidak, jawab `animal` atau `unknown_animal`, bukan mengarang spesies.

Dengan cara ini sistem dapat memiliki vocabulary luas tetapi coverage detail tetap jujur.

## 9. Ekspresi dan aktivitas

Ekspresi tidak dapat diselesaikan hanya dengan ontology. Ia membutuhkan face crop berkualitas, landmark, pose, dan label manusia yang konsisten. Aktivitas bahkan memerlukan temporal evidence jika berasal dari video; satu foto sering hanya memberi snapshot ambigu.

Untuk face specialist:

```text
face_quality = face_size × sharpness × visibility × landmark_stability
```

Jika `face_quality < τ_face`, jangan menebak ekspresi. Untuk aktivitas:

```text
activity_confidence = object_relation × pose_evidence × context × visibility
```

`driving` memerlukan person, vehicle, pose, dan kontak spatial yang mendukung; satu mobil dan satu orang dalam frame tidak cukup.

## 10. Apa yang benar-benar dipelajari dari DeepSeek?

Speculative decoding menunjukkan pola umum yang sangat berguna:

| Prinsip | Transfer ke HeraVision |
|---|---|
| Draft murah | Region/prototype proposal P0/P1 |
| Target verifier | Evidence + specialist + knowledge constraints |
| Acceptance rate | Persentase proposal yang lolos verifikasi |
| Confidence scheduling | Tentukan kedalaman verifikasi per sample |
| Early rejection | Buang proposal lemah sedini mungkin |
| Benchmark bottleneck | Ukur latency proposal dan verifier terpisah |

Namun ada hal yang **tidak boleh disamakan**. Speculative decoding dapat mempertahankan distribusi target karena target model memverifikasi token [1] [2]. HeraVision tidak memiliki target semantic universal. Karena itu, semantic answer tetap bersifat selective dan dapat abstain.

DeepSpec sendiri membutuhkan target cache besar dan konfigurasi default multi-GPU [3]. Ini justru menguatkan pelajaran bahwa efisiensi harus diukur pada bottleneck aktual; kita tidak dapat menyalin stack DeepSeek ke PC kentang dan menganggap hasilnya sama.

## 11. Eksperimen penemuan HeraVision

Hipotesis yang layak dipublikasikan adalah:

> **A confidence-scheduled sparse visual cascade can reduce average CPU cost by routing only ambiguous regions to expensive specialists, while preserving evidence completeness and reducing unsafe semantic answers.**

Uji minimal:

| Eksperimen | Ukuran utama | Syarat berhasil |
|---|---|---|
| Always-P2 vs cascade | p95 latency, RAM | cascade lebih murah tanpa F1 turun tajam |
| Fixed verification vs scheduled | average cost | scheduled cost turun dengan unsafe rate tidak naik |
| All classes vs top-k retrieval | latency | retrieval lebih cepat dengan recall broad setara |
| No margin vs margin gate | unsafe rate | margin menurunkan jawaban ambigu |
| Full image vs stable region | CPU time | region route lebih murah dengan evidence tetap lengkap |

Semua eksperimen perlu development split, holdout split, dan blind retest. Dengan 30–60 gambar kita hanya dapat menguji plumbing dan failure modes. Kita belum dapat menyimpulkan kemampuan “semua hewan” atau “semua ekspresi”.

## Roadmap implementasi

```text
L1 ontology + broad prototype bank
→ top-k retrieval dari region B1
→ evidence-aware scorer
→ margin + abstention gate
→ confidence-scheduled specialist
→ face micro-benchmark
→ holdout multi-domain
→ vocabulary expansion
```

## Kesimpulan

Celahnya ada pada **representasi, routing, verifikasi, dan reuse**, bukan pada menghapus matematika yang diperlukan untuk mengenali objek. Kita dapat membuat PC kentang menjadi jauh lebih efisien dengan tidak menjalankan pengetahuan dan specialist yang tidak relevan.

HeraVision dapat menjadi sangat luas pengetahuannya jika ontology dan prototype bank-nya luas, tetapi ia harus tetap mengakui bahwa **knowledge coverage tidak sama dengan visual recognition coverage**. Penemuan yang paling realistis adalah mesin sparse yang menghabiskan komputasi hanya pada bukti yang bernilai, memverifikasi proposal murah, dan abstain ketika dunia visual tidak cukup jelas.

## References

[1]: https://research.google/blog/looking-back-at-speculative-decoding/ "Google Research: Looking back at speculative decoding"
[2]: https://arxiv.org/html/2402.01528v3 "Decoding Speculative Decoding"
[3]: https://github.com/deepseek-ai/DeepSpec "DeepSeek AI DeepSpec repository"
[4]: https://arxiv.org/abs/2607.05147 "DSpark: Confidence-Scheduled Speculative Decoding with Semi-Autoregressive Generation"
