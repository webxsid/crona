import { EventSource } from "eventsource";

class EventStack {

  private events: {
    event: string;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    data: any;
  }[] = [];

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  push(event: string, data: any) {
    this.events.push({ event, data });
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  getEvents(): { event: string; data: any }[] {
    return this.events;
  }

  clear() {
    this.events = [];
  }
}

export function listenEvents(baseUrl: string, token: string): {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  events: () => { event: string; data: any }[];
  close: () => void;
} {

  const events = new EventStack();
  const es = new EventSource(`${baseUrl}/events`, {
    fetch: (url, init) => {
      init.headers = {
        ...init.headers,
        Authorization: `Bearer ${token}`
      };
      return fetch(url, init);
    }
  });

  // es.onopen = () => {
  //   console.log("SSE connected");
  // };

  es.onerror = (err) => {
    console.log("SSE connection error", {
      baseUrl,
      token
    });
    console.error("SSE error", err);
  };

  es.addEventListener("timer.state", (e) => {
    events.push("timer.state", JSON.parse((e as MessageEvent).data));
  });

  es.addEventListener("context.changed", (e) => {
    events.push("context.changed", JSON.parse((e as MessageEvent).data));
  })

  es.addEventListener("session.started", (e) => {
    events.push("session.started", JSON.parse((e as MessageEvent).data));
  })

  es.addEventListener("session.stopped", (e) => {
    events.push("session.stopped", JSON.parse((e as MessageEvent).data));
  })

  es.addEventListener("timer.boundary", (e) => {
    events.push("timer.boundary", JSON.parse((e as MessageEvent).data));
  })

  es.addEventListener("stash.created", (e) => {
    events.push("stash.created", JSON.parse((e as MessageEvent).data));
  })

  es.addEventListener("stash.applied", (e) => {
    events.push("stash.applied", JSON.parse((e as MessageEvent).data));
  })


  es.addEventListener("stash.dropped", (e) => {
    events.push("stash.dropped", JSON.parse((e as MessageEvent).data));
  })

  es.addEventListener("repo.created", (e) => {
    events.push("repo.created", JSON.parse((e as MessageEvent).data));
  })

  // catch all events
  es.onmessage = (e) => {
    if (e.data === ":ok") {
      // ignore initial ping
      return;
    }
    events.push("message", JSON.parse(e.data));
  }

  return {
    events: () => events.getEvents(),
    close: () => {
      events.clear();
      es.close();
    }
  };
}
