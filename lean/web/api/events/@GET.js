function handler (w, r) {
    const lastEventId = r.header.get("last-event-id")
    const nextEvent = streamEvents(r.context(),lastEventId)
    sendServerEvents(() => {
      const evt = nextEvent()
      return { event: 'event', data: evt.event, id: evt.id }
    })
  }