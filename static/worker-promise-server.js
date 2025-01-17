const isPromise = (obj) => {
  return (
    !!obj &&
    (typeof obj === "object" || typeof obj === "function") &&
    typeof obj.then === "function"
  );
};

const registerPromiseWorker = (callback) => {
  const postOutgoingMessage = (e, messageId, error, result) => {
    const postMessage = (msg) => {
      self.postMessage(msg);
    };

    if (error) {
      console.error("Worker caught an error:", error);
      postMessage([
        messageId,
        {
          message: error.message,
        },
      ]);
    } else {
      postMessage([messageId, null, result]);
    }
  };

  const tryCatchFunc = (callback, message) => {
    try {
      return { res: callback(message) };
    } catch (e) {
      return { err: e };
    }
  };

  const handleIncomingMessage = (e, callback, messageId, message) => {
    const result = tryCatchFunc(callback, message);

    if (result.err) {
      postOutgoingMessage(e, messageId, result.err);
    } else if (!isPromise(result.res)) {
      postOutgoingMessage(e, messageId, null, result.res);
    } else {
      result.res
        .then((finalResult) => {
          postOutgoingMessage(e, messageId, null, finalResult);
        })
        .catch((finalError) => {
          postOutgoingMessage(e, messageId, finalError);
        });
    }
  };

  const onIncomingMessage = (e) => {
    const { data: payload } = e;

    if (!Array.isArray(payload) || payload.length !== 2) {
      // message doesn't match communication format; ignore
      return;
    }

    const [messageId, message] = payload;

    if (typeof callback !== "function") {
      postOutgoingMessage(
        e,
        messageId,
        new Error("Please pass a function into register()."),
      );
    } else {
      handleIncomingMessage(e, callback, messageId, message);
    }
  };

  self.addEventListener("message", onIncomingMessage);
};
