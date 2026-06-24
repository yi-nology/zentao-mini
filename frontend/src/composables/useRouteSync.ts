import { useRoute, useRouter, type LocationQueryRaw } from 'vue-router'

type Primitive = string | number | boolean | null | undefined
type QueryValue = Primitive | Primitive[]

function isEmpty(v: QueryValue): boolean {
  return v === null || v === undefined || v === ''
}

function serialize(val: QueryValue): string | string[] | undefined {
  if (isEmpty(val)) return undefined
  if (Array.isArray(val)) {
    const arr = val.filter(v => !isEmpty(v)).map(String)
    return arr.length ? arr : undefined
  }
  return String(val)
}

export function buildQuery(
  params: Record<string, QueryValue>,
  defaults: Record<string, QueryValue> = {}
): LocationQueryRaw {
  const q: LocationQueryRaw = {}
  for (const [key, val] of Object.entries(params)) {
    const def = defaults[key]
    if (def !== undefined && String(def) === String(val)) continue
    const s = serialize(val)
    if (s !== undefined) q[key] = s
  }
  return q
}

export function readQuery<T extends Record<string, QueryValue>>(
  defaults: T
): T {
  const route = useRoute()
  const result = { ...defaults } as Record<string, QueryValue>
  for (const [key, def] of Object.entries(defaults)) {
    const qv = route.query[key]
    if (qv === undefined) continue
    if (Array.isArray(def)) {
      result[key] = Array.isArray(qv) ? qv : [qv]
    } else if (typeof def === 'number') {
      const n = Number(qv)
      if (!isNaN(n)) result[key] = n
    } else {
      result[key] = Array.isArray(qv) ? qv[0] : qv
    }
  }
  return result as T
}

export function useRouteSync() {
  const router = useRouter()

  function updateQuery(params: Record<string, QueryValue>, defaults: Record<string, QueryValue> = {}) {
    const q = buildQuery(params, defaults)
    router.replace({ query: q })
  }

  return { updateQuery, buildQuery, readQuery }
}
