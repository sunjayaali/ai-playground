/**
 * Realtime visitor counter (NuxtHub realtime guide).
 *
 * Peers subscribe to the `visitors` topic on connect and publish their
 * count to it, so every open socket is told how many peers are online.
 * On disconnect we unsubscribe and republish after a beat, giving crossws
 * a moment to prune the closed peer before the surviving clients read
 * `peer.peers.size` — otherwise they would briefly see a stale count.
 *
 * `peer.publish` fans out to every subscriber of the topic, including
 * the publishing peer itself.
 */
export default defineWebSocketHandler({
  open(peer) {
    peer.subscribe("visitors");
    peer.publish("visitors", peer.peers.size);
    peer.send(peer.peers.size);
  },
  close(peer) {
    peer.unsubscribe("visitors");
    setTimeout(() => {
      peer.publish("visitors", peer.peers.size);
    }, 500);
  },
});
