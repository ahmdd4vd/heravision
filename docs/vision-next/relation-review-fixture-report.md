# MD1 Relation Review Fixture

**Fixture:** `experiments/manifests/md1-relation-review-fixture.json`  
**Samples:** 30 across Imagenette, Imagewoof, and Wikimedia diagrams  
**Predicted edges for review:** 50  
**Status:** pending independent relation review

This fixture extracts B1 stable visible relation edges together with their endpoint boxes. An independent reviewer may mark each edge `correct`, `incorrect`, or `uncertain` based on the original image. The fixture measures predicted-edge precision and false relation rate only. It cannot measure missed relations because unpredicted edges are not present; a separate relation-complete annotation is required for recall.

The fixture deliberately preserves the model's predicate and provenance context while leaving every review decision blank. It must not be used as a benchmark until an independent reviewer completes the decisions.
