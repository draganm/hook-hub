function handler (w, r) {
    const nextEvent = streamEvents(r.context(),"")
    sendServerEvents(() => {
      const evt = nextEvent()
      return { event: 'event', data: JSON.stringify(evt) }
    })
  }