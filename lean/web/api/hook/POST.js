function handler(w,r) {
    storeEvent(readToString(r.body))
}