function handler(w,r) {
    if (!isAccessTokenValid(r)) {
        returnStatus(403, "not authenticated")
    }
    w.write("hello world!")
}