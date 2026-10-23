// Speaker / microphone through the ES8311 codec. Playback runs in its own
// task, so the UI never waits for audio.
#pragma once

#include "esp_err.h"

typedef enum {
    AUDIO_FX_TAP,      // short click
    AUDIO_FX_CORRECT,  // two rising notes
    AUDIO_FX_WRONG,    // low buzz
    AUDIO_FX_CHIME,    // study reminder
} audio_fx_t;

esp_err_t audio_init(void);
void audio_set_volume(int volume); // 0..100
int audio_get_volume(void);
void audio_fx(audio_fx_t fx);

/** Plays a WAV (16-bit PCM, mono or stereo) from http(s), in the background. */
void audio_play_url(const char *url);
void audio_stop(void);
