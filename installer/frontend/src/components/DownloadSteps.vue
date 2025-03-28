<script setup lang="ts">
import { canContinue, step } from '../global';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import { ref, watch } from 'vue';
import { GetPin, FinishRegistration, DownloadPlaydateOS } from '../../wailsjs/go/main/App'

const error = ref('');

const serialNumber = ref('PDU1-Y');
const pin = ref('');

let accessToken: string;

const downloaded = ref(false);

watch(step, async (newStep, oldStep) => {
  error.value = '';

  if (oldStep === 2 && newStep === 3) {
    if (!/^PDU1-Y\d{6}$/g.test(serialNumber.value)) {
      step.value--;
      return;
    }
    const pinInfo = await GetPin(serialNumber.value);
    if (pinInfo.pin === undefined) {
      step.value--;
      if (pinInfo.detail !== undefined) error.value = pinInfo.detail;
      return;
    }
    pin.value = pinInfo.pin;
    return;
  }
  if (oldStep === 3 && newStep === 4) {
    canContinue.value = false;
    const info = await FinishRegistration(serialNumber.value);
    if (info["access_token"] === undefined) {
      step.value--;
      if (info.detail !== undefined) error.value = info.detail;
      if (info.registered !== undefined && !info.registered) error.value = "The pin was not entered."
      return;
    }
    accessToken = info["access_token"];

    await DownloadPlaydateOS(accessToken);
    canContinue.value = true;
    downloaded.value = true;
  }
});
</script>
<template>
  <span class="error" v-if="error != ''">Error: {{ error }}</span>
  <template v-if="step === 1">
    <p>
      Please go to https://play.date/devices/, select your Playdate, and remove it from your account. Press next when
      your done.
    </p>
    <button @click="BrowserOpenURL('https://play.date/devices/')">Open in Browser</button>
  </template>
  <template v-else-if="step === 2">
    <p>
      Please input your Playdate's serial number.
    </p>
    <input v-model="serialNumber" placeholder="PDU1-Y" type="text" />
  </template>
  <template v-else-if="step === 3">
    <p>Please go to https://play.date/pin and enter the following pin: <b>{{ pin }}</b></p>
    <button @click="BrowserOpenURL('https://play.date/pin/' + pin)">Open in Browser</button>
  </template>
  <template v-else-if="step === 4">
    <template v-if="downloaded">
      <p>Successfully download PlaydateOS!</p>
    </template>
    <template v-else>
      <p>Downloading PlaydateOS...</p>
      <div class="loader" />
    </template>
  </template>
</template>
