# Abstention Policy v2

## Tujuan

HeraVision harus memilih antara menjawab, abstain, dan menyatakan evidence tidak cukup. Status tersebut adalah keputusan keamanan klaim, bukan label objek dan bukan probabilitas terkalibrasi.

## Aturan keputusan

| Status | Syarat visual minimum | Tindakan |
|---|---|---|
| `answered` | Region stabil pada minimal dua skala, boundary cukup kuat, dan tidak ada konflik scope besar | Hanya keluarkan klaim generic berbasis evidence |
| `abstain` | Ada region tetapi stability/boundary lemah, objek saling overlap, atau scope komposisi ambigu | Jangan keluarkan klaim positif; simpan warning dan evidence |
| `insufficient_evidence` | Tidak ada region yang lolos filter atau input tidak menyediakan struktur yang dapat digunakan | Nyatakan evidence tidak cukup |

## Guardrail

HeraVision tidak boleh mengubah status `answered` menjadi label semantic seperti `dog`, `person`, atau `car`. Status `answered` pada fase ini hanya berarti struktur visual generic terdeteksi.

Relasi `touching` tidak boleh dipakai sebagai evidence kuat hanya karena dua bbox overlap. Relasi tersebut harus memiliki boundary-gap evidence pada pixel atau diturunkan menjadi `uncertain`.

## Candidate threshold

The current conservative candidate threshold is `answer-min-score=0.65`. At this threshold, the same-agent internal fixture showed 75% coverage on internally expected-answerable images and 33.33% unsafe-answer rate on internally expected-nonanswerable images. The baseline threshold 0.45 showed 79.17% coverage and 83.33% unsafe-answer rate. This is a development trade-off, not calibrated probability.

## Baseline internal

Pada 30 sample MD1, blind third internal review menandai 24 `answered`, 5 `abstain`, dan 1 `insufficient_evidence`. Output B1 stable saat ini menghasilkan coverage 79,17% pada internal expected-answerable set dan unsafe-answer rate 83,33% pada internal expected-nonanswerable set.

Angka tersebut adalah development diagnostic dari same-agent review, bukan independent human benchmark. Threshold baru hanya boleh dipromosikan jika unsafe-answer rate turun tanpa membuat coverage answerable runtuh pada fixture yang sama dan blind retest tetap bebas error.
