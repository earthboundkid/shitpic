let messageIds = 0;

function onMessage(self, e) {
  const message = e.data;
  if (!Array.isArray(message) || message.length < 2) {
    // Ignore - this message is not for us.
    return;
  }

  const [messageId, error, result] = message;
  const callback = self._callbacks[messageId];

  if (!callback) {
    // Ignore - user might have created multiple PromiseWorkers.
    // This message is not for us.
    return;
  }

  self._callbacks[messageId] = undefined;
  callback(error, result);
}

class PromiseWorker {
  constructor(worker) {
    this._worker = worker;
    this._callbacks = {};

    worker.addEventListener("message", (e) => {
      onMessage(this, e);
    });
  }

  postMessage(userMessage) {
    const messageId = messageIds++;
    const messageToSend = [messageId, userMessage];

    return new Promise((resolve, reject) => {
      this._callbacks[messageId] = (error, result) => {
        if (error) {
          return reject(error);
        }
        resolve(result);
      };

      this._worker.postMessage(messageToSend);
    });
  }
}

export default PromiseWorker;
