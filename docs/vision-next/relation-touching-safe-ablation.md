# Relation Touching-Safe Ablation

## Change

The opt-in `-relation-touching-safe` mode suppresses `touching` when one endpoint bbox contains the other. It preserves touching for boundary-adjacent boxes and emits `boundary-contact` evidence. This avoids treating containment as pixel contact.

## MD1 candidate result

| Configuration | Total edges | Contains | Overlapping | Touching | Above | Left_of |
|---|---:|---:|---:|---:|---:|---:|
| Relation prune baseline | 50 | 12 | 16 | 17 | 3 | 2 |
| Touching-safe | 38 | 12 | 16 | 5 | 3 | 2 |

All 30 samples completed without errors. The change removes 12 containment-derived touching edges while preserving all other predicate counts in this fixture. The mode remains a candidate until an independent relation-complete annotation confirms that the removed edges were false and the retained touching edges are visually supported.
