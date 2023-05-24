/**
 * @module skottie-sk
 * @description <h2><code>skottie-sk</code></h2>
 *
 * <p>
 *   The main application element for skottie.
 * </p>
 *
 */
import '../skottie-config-sk';
import '../skottie-player-sk';
import '../../../elements-sk/modules/checkbox-sk';
import '../../../elements-sk/modules/collapse-sk';
import '../../../elements-sk/modules/error-toast-sk';
import { html, TemplateResult } from 'lit-html';
import JSONEditor from 'jsoneditor';
import LottiePlayer from 'lottie-web';
import { $$ } from '../../../infra-sk/modules/dom';
import { errorMessage } from '../../../elements-sk/modules/errorMessage';
import { define } from '../../../elements-sk/modules/define';
import { jsonOrThrow } from '../../../infra-sk/modules/jsonOrThrow';
import { stateReflector } from '../../../infra-sk/modules/stateReflector';
import { CollapseSk } from '../../../elements-sk/modules/collapse-sk/collapse-sk';
import { SkottieGifExporterSk } from '../skottie-gif-exporter-sk/skottie-gif-exporter-sk';
import '../skottie-gif-exporter-sk';
import '../skottie-text-editor-sk';
import { replaceTexts } from '../skottie-text-editor-sk/text-replace';
import '../skottie-library-sk';
import { SoundMap, AudioPlayer } from '../audio';
import '../skottie-performance-sk';
import { renderByDomain } from '../helpers/templates';
import { supportedDomains } from '../helpers/domains';
import '../skottie-audio-sk';
import { ElementSk } from '../../../infra-sk/modules/ElementSk';
import {
  SkottieConfigEventDetail,
  SkottieConfigState,
} from '../skottie-config-sk/skottie-config-sk';
import { SkottiePlayerSk } from '../skottie-player-sk/skottie-player-sk';
import { SkottiePerformanceSk } from '../skottie-performance-sk/skottie-performance-sk';
import { FontAsset, LottieAnimation, LottieAsset, ViewMode } from '../types';
import { SkottieLibrarySk } from '../skottie-library-sk/skottie-library-sk';
import {
  AudioStartEventDetail,
  SkottieAudioSk,
} from '../skottie-audio-sk/skottie-audio-sk';
import {
  SkottieTextEditorSk,
  TextEditApplyEventDetail,
} from '../skottie-text-editor-sk/skottie-text-editor-sk';
import '../skottie-shader-editor-sk';
import {
  ShaderEditApplyEventDetail,
  ShaderEditorSk,
} from '../skottie-shader-editor-sk/skottie-shader-editor-sk';
import '../../../infra-sk/modules/theme-chooser-sk';
import '../../../infra-sk/modules/app-sk';
import { replaceShaders } from '../skottie-shader-editor-sk/shader-replace';
import '../../../elements-sk/modules/icons/expand-less-icon-sk';
import '../../../elements-sk/modules/icons/expand-more-icon-sk';
import '../../../elements-sk/modules/icons/play-arrow-icon-sk';
import '../../../elements-sk/modules/icons/pause-icon-sk';
import '../../../elements-sk/modules/icons/replay-icon-sk';
import '../skottie-button-sk';
import '../skottie-dropdown-sk';
import { DropdownSelectEvent } from '../skottie-dropdown-sk/skottie-dropdown-sk';
import '../skottie-exporter-sk';
import {
  ExportType,
  SkottieExporterSk,
} from '../skottie-exporter-sk/skottie-exporter-sk';
import '../skottie-file-settings-sk';
import { SkottieFileSettingsEventDetail } from '../skottie-file-settings-sk/skottie-file-settings-sk';

// It is assumed that this symbol is being provided by a version.js file loaded in before this
// file.
declare const SKIA_VERSION: string;

interface BodymovinPlayer {
  goToAndStop(t: number): void;
  goToAndPlay(t: number): void;
  pause(): void;
  play(): void;
}

interface LottieLibrary {
  version: string;
  loadAnimation(opts: Record<string, unknown>): BodymovinPlayer;
}

interface LoadedAsset {
  name: string;
  bytes?: ArrayBuffer;
  player?: AudioPlayer;
}

const GOOGLE_WEB_FONTS_HOST =
  'https://storage.googleapis.com/skia-cdn/google-web-fonts';

const PRODUCTION_ASSETS_PATH = '/_/a';

// Make this the hash of the lottie file you want to play on startup.
const DEFAULT_LOTTIE_FILE = '1112d01d28a776d777cebcd0632da15b'; // gear.json

// SCRUBBER_RANGE is the input range for the scrubbing control.
// This is an arbitrary value, and is treated as a re-scaled duration.
const SCRUBBER_RANGE = 1000;

const AUDIO_SUPPORTED_DOMAINS = [
  supportedDomains.SKOTTIE_INTERNAL,
  supportedDomains.SKOTTIE_TENOR,
  supportedDomains.LOCALHOST,
];

type UIMode = 'dialog' | 'loading' | 'loaded';

const caption = (text: string, mode: ViewMode) => {
  if (mode === 'presentation') {
    return null;
  }
  return html` <figcaption>${text}</figcaption> `;
};

const redir = () =>
  renderByDomain(
    html` <div>
      Googlers should use
      <a href="https://skottie-internal.skia.org">skottie-internal.skia.org</a>.
    </div>`,
    Object.values(supportedDomains).filter(
      (domain: string) => domain !== supportedDomains.SKOTTIE_INTERNAL
    )
  );

const displayLoading = () => html`
  <div class="loading">
    <spinner-sk active></spinner-sk><span>Loading...</span>
  </div>
`;

export class SkottieSk extends ElementSk {
  private static template = (ele: SkottieSk) => html`
    <app-sk>
      <div class="app-container">
        <header>
          <h2>Skottie Web Player</h2>
          <span>
            <a
              href="https://skia.googlesource.com/skia/+show/${SKIA_VERSION}"
              class="header__skia-version"
            >
              ${SKIA_VERSION.slice(0, 7)}
            </a>

            <skottie-dropdown-sk
              id="view-exporter"
              .name="dropdown-exporter"
              .options=${[
                { id: '', value: 'Export' },
                { id: 'gif', value: 'GIF' },
                { id: 'webM', value: 'WebM' },
                { id: 'png', value: 'PNG sequence' },
              ]}
              reset
              @select=${ele.exportSelectHandler}
              border
            >
            </skottie-dropdown-sk>
            <skottie-button-sk
              id="view-perf-chart"
              @select=${ele.togglePerformanceChart}
              type="outline"
              .content=${'Performance chart'}
              .classes=${['header__button']}
            >
            </skottie-button-sk>
            <skottie-button-sk
              id="view-json-layers"
              @select=${ele.toggleEditor}
              type="outline"
              .content=${'View JSON code'}
              .classes=${['header__button']}
            >
            </skottie-button-sk>

            <theme-chooser-sk></theme-chooser-sk>
          </span>
        </header>
        <main>${ele.pick()}</main>
        <footer>
          <error-toast-sk></error-toast-sk>
          ${redir()}
        </footer>
      </div>
      <skottie-exporter-sk @start=${ele.onExportStart}></skottie-exporter-sk>
    </app-sk>
  `;

  // pick the right part of the UI to display based on ele._ui.
  private pick = () => {
    switch (this.ui) {
      default:
      case 'dialog':
        return this.displayDialog();
      case 'loading':
        return displayLoading();
      case 'loaded':
        return this.displayLoaded();
    }
  };

  private displayDialog = () => html`
    <skottie-config-sk
      .state=${this.state}
      .width=${this.width}
      .height=${this.height}
      .fps=${this.fps}
      .backgroundColor=${this.backgroundColor}
      @skottie-selected=${this.skottieFileSelected}
      @cancelled=${this.selectionCancelled}
    ></skottie-config-sk>
  `;

  private displayLoaded = () => html`
    <div class="threecol">
      <div class="left">${this.leftControls()}</div>
      <div class="main">${this.mainContent()}</div>
      <div class="right">${this.rightControls()}</div>
    </div>
  `;

  private mainContent = () => html`
    <div class="players">
      <figure class="players-container">
        ${this.skottiePlayerTemplate()} ${this.lottiePlayerTemplate()}
      </figure>
      ${this.livePreview()}
    </div>
    <div class="playback">
      <div class="playback-content">
        <skottie-button-sk
          id="playpause"
          .content=${html`<play-arrow-icon-sk
              id="playpause-play"
            ></play-arrow-icon-sk>
            <pause-icon-sk id="playpause-pause"></pause-icon-sk>`}
          .classes=${['playback-content__button']}
          @select=${this.playpause}
        ></skottie-button-sk>
        <div class="scrub">
          <input
            id="scrub"
            type="range"
            min="0"
            max=${SCRUBBER_RANGE}
            step="0.1"
            @input=${this.onScrub}
            @change=${this.onScrubEnd}
          />
          <label class="number">
            Frame:
            <input
              type="number"
              id="frameInput"
              class="playback-content-frameInput"
              @focus=${this.onFrameFocus}
              @change=${this.onFrameChange}
            /><!--
            --><span class="playback-content-frameTotal" id="frameTotal"
              >of 0</span
            >
          </label>
        </div>
        <skottie-button-sk
          id="rewind"
          .content=${html`<replay-icon-sk></replay-icon-sk>`}
          .classes=${['playback-content__button']}
          @select=${this.rewind}
        ></skottie-button-sk>
      </div>
    </div>

    <div @click=${this.onChartClick}>${this.performanceChartTemplate()}</div>
    ${this.jsonEditor()} ${this.gifExporter()}

    <collapse-sk id="volume" closed>
      <p>Volume:</p>
      <input
        id="volume-slider"
        type="range"
        min="0"
        max="1"
        step=".05"
        value="1"
        @input=${this.onVolumeChange}
      />
    </collapse-sk>
  `;

  private embedDialog() {
    return html`
      <details class="embed expando">
        <summary id="embed-open">
          <span>Embed</span><expand-less-icon-sk></expand-less-icon-sk
          ><expand-more-icon-sk></expand-more-icon-sk>
        </summary>
        <label>
          Embed using an iframe
          <input value=${this.iframeDirections()} />
        </label>
        <label>
          Embed on skia.org
          <input value=${this.inlineDirections()} />
        </label>
      </details>
    `;
  }

  private performanceChartTemplate() {
    return html`
      <dialog class="perf-chart" ?open=${this.showPerformanceChart}>
        <div class="top-ribbon">
          <span>Performance Chart</span>
          <button @click=${this.togglePerformanceChart}>Close</button>
        </div>
        <skottie-performance-sk id="chart"></skottie-performance-sk>
      </dialog>
    `;
  }

  private leftControls = () => {
    if (this.viewMode === 'presentation') {
      return null;
    }

    return html`
      <div class="json-chooser">
        <div class="title">JSON File</div>
        <div class="upload-download">
          <button class="edit-config large" @click=${this.startEdit}>
            ${this.state.filename} ${this.width}x${this.height} ...
          </button>
          <div class="download">
            <a
              target="_blank"
              download=${this.state.filename}
              href=${this.downloadURL}
            >
              Download
            </a>
            ${this.hasEdits ? '(without edits)' : ''}
          </div>
        </div>
      </div>

      ${this.fileSettingsDialog()} ${this.audioDialog()} ${this.optionsDialog()}

      <button
        class="apply-button"
        ?hidden=${!this.hasEdits}
        @click=${this.applyEdits}
      >
        Apply Edits
      </button>
    `;
  };

  private rightControls = () => html`
    ${this.jsonTextEditor()} ${this.library()} ${this.embedDialog()}
  `;

  private optionsDialog = () => html`
    <details class="expando">
      <summary id="options-open">
        <span>Options</span><expand-less-icon-sk></expand-less-icon-sk
        ><expand-more-icon-sk></expand-more-icon-sk>
      </summary>
      <div class="options-container">
        <checkbox-sk
          label="Show lottie-web"
          ?checked=${this.showLottie}
          @click=${this.toggleLottie}
        >
        </checkbox-sk>
      </div>
    </details>
  `;

  private audioDialog = () =>
    renderByDomain(
      html`
        <details
          class="expando"
          ?open=${this.showAudio}
          @toggle=${(e: Event) =>
            this.toggleAudio((e.target! as HTMLDetailsElement).open)}
        >
          <summary id="audio-open">
            <span>Audio</span><expand-less-icon-sk></expand-less-icon-sk
            ><expand-more-icon-sk></expand-more-icon-sk>
          </summary>

          <skottie-audio-sk
            .animation=${this.state.lottie}
            @apply=${this.applyAudioSync}
          >
          </skottie-audio-sk>
        </details>
      `,
      AUDIO_SUPPORTED_DOMAINS
    );

  private fileSettingsDialog = () =>
    html`
      <details
        class="expando"
        ?open=${this.showFileSettings}
        @toggle=${(e: Event) =>
          this.toggleFileSettings((e.target! as HTMLDetailsElement).open)}
      >
        <summary id="fileSettings-open">
          <span>File Settings</span><expand-less-icon-sk></expand-less-icon-sk
          ><expand-more-icon-sk></expand-more-icon-sk>
        </summary>
        <skottie-file-settings-sk
          .width=${this.width}
          .height=${this.height}
          .fps=${this.fps}
          @settings-change=${this.skottieFileSettingsUpdated}
        ></skottie-file-settings-sk>
      </details>
    `;

  private iframeDirections = () =>
    `<iframe width="${this.width}" height="${this.height}" src="${window.location.origin}/e/${this.hash}?w=${this.width}&h=${this.height}" scrolling=no>`;

  private inlineDirections = () =>
    `<skottie-inline-sk width="${this.width}" height="${this.height}" src="${window.location.origin}/_/j/${this.hash}"></skottie-inline-sk>`;

  private skottiePlayerTemplate = () => html` <figure
    class="players-container-player"
  >
    <skottie-player-sk paused width=${this.width} height=${this.height}>
    </skottie-player-sk>
    ${this.wasmCaption()}
  </figure>`;

  private lottiePlayerTemplate = () => {
    if (!this.showLottie) {
      return '';
    }
    return html` <figure class="players-container-player">
      <div
        id="container"
        title="lottie-web"
        style="width: 100%; aspect-ratio: ${this.width /
        this.height}; background-color: ${this.backgroundColor}"
      ></div>
      ${caption('lottie-web', this.viewMode)}
    </figure>`;
  };

  private library = () => html` <details
    class="expando"
    ?open=${this.showLibrary}
    @toggle=${(e: Event) =>
      this.toggleLibrary((e.target! as HTMLDetailsElement).open)}
  >
    <summary id="library-open">
      <span>Library</span><expand-less-icon-sk></expand-less-icon-sk
      ><expand-more-icon-sk></expand-more-icon-sk>
    </summary>

    <skottie-library-sk @select=${this.updateAnimation}> </skottie-library-sk>
  </details>`;

  // TODO(kjlubick): Make the live preview use skottie
  private livePreview = () => {
    if (!this.hasEdits || !this.showLottie) {
      return '';
    }
    if (this.hasEdits) {
      return html` <figure>
        <div
          id="live"
          title="live-preview"
          style="width: ${this.width}px; height: ${this.height}px"
        ></div>
        <figcaption>Preview [lottie-web]</figcaption>
      </figure>`;
    }
    return '';
  };

  private jsonEditor = (): TemplateResult => html` <dialog
    class="editor"
    ?open=${this.showJSONEditor}
  >
    <div class="top-ribbon">
      <span>Layer Information</span>
      <button @click=${this.toggleEditor}>Close</button>
    </div>
    <div id="json_editor"></div>
  </dialog>`;

  private gifExporter = () => html`
    <dialog class="export" ?open=${this.showGifExporter}>
      <div class="top-ribbon">
        <span>Export</span>
        <button @click=${this.toggleGifExporter}>Close</button>
      </div>
      <skottie-gif-exporter-sk @start=${this.onExportStart}>
      </skottie-gif-exporter-sk>
    </dialog>
  `;

  private jsonTextEditor = () => html`
    <details
      class="expando"
      ?open=${this.showTextEditor}
      @toggle=${(e: Event) =>
        this.toggleTextEditor((e.target! as HTMLDetailsElement).open)}
    >
      <summary id="edit-text-open">
        <span>Edit Text</span><expand-less-icon-sk></expand-less-icon-sk
        ><expand-more-icon-sk></expand-more-icon-sk>
      </summary>

      <skottie-text-editor-sk
        .animation=${this.state.lottie}
        .mode=${this.viewMode}
        @apply=${this.applyTextEdits}
      >
      </skottie-text-editor-sk>
    </details>
  `;

  private shaderEditor = () => html`
    <details
      class="expando"
      ?open=${this.showShaderEditor}
      @toggle=${(e: Event) =>
        this.toggleShaderEditor((e.target! as HTMLDetailsElement).open)}
    >
      <summary>
        <span>Edit Shader</span><expand-less-icon-sk></expand-less-icon-sk
        ><expand-more-icon-sk></expand-more-icon-sk>
      </summary>

      <skottie-shader-editor-sk
        .animation=${this.state.lottie}
        .mode=${this.viewMode}
        @apply=${this.applyShaderEdits}
      >
      </skottie-shader-editor-sk>
    </details>
  `;

  private buildFileName = () => {
    const fileName =
      this.state.filename || this.state.lottie?.metadata?.filename;
    if (fileName) {
      return html`<div title="${fileName}">${fileName}</div>`;
    }
    return null;
  };

  private wasmCaption = () => {
    if (this.viewMode === 'presentation') {
      return null;
    }
    return html` <figcaption style="max-width: ${this.width}px;">
      <div>skottie-wasm</div>
      ${this.buildFileName()}
    </figcaption>`;
  };

  private assetsPath = PRODUCTION_ASSETS_PATH; // overridable for testing

  // The URL referring to the lottie JSON Blob.
  private backgroundColor: string = 'rgba(0,0,0,0)';

  private downloadURL: string = '';

  private duration: number = 0; // 0 is a sentinel value for "player not loaded yet"

  private editor: JSONEditor | null = null;

  private editorLoaded: boolean = false;

  // used for remembering the time elapsed while the animation is playing.
  private elapsedTime: number = 0;

  private fps: number = 0;

  private hasEdits: boolean = false;

  private hash: string = '';

  private height: number = 0;

  private live: BodymovinPlayer | null = null;

  private lottiePlayer: BodymovinPlayer | null = null;

  private performanceChart: SkottiePerformanceSk | null = null;

  private playing: boolean = true;

  private playingOnStartOfScrub: boolean = false;

  // The wasm animation computes how long it has been since the previous rendered time and
  // uses arithmetic to figure out where to seek (i.e. which frame to draw).
  private previousFrameTime: number = 0;

  private scrubbing: boolean = false;

  private showAudio: boolean = false;

  private showGifExporter: boolean = false;

  private showJSONEditor: boolean = false;

  private showLibrary: boolean = false;

  private showLottie: boolean = false;

  private showPerformanceChart: boolean = false;

  private showTextEditor: boolean = false;

  private showShaderEditor: boolean = false;

  private showFileSettings: boolean = false;

  private skottieLibrary: SkottieLibrarySk | null = null;

  private skottiePlayer: SkottiePlayerSk | null = null;

  private speed: number = 1; // this is a playback multiplier

  private state: SkottieConfigState;

  private stateChanged: () => void;

  private ui: UIMode = 'dialog';

  private viewMode: ViewMode = 'default';

  private width: number = 0;

  constructor() {
    super(SkottieSk.template);

    this.state = {
      filename: '',
      lottie: null,
      assetsZip: '',
      assetsFilename: '',
    };

    this.stateChanged = stateReflector(
      /* getState */ () => ({
        // provide empty values
        l: this.showLottie,
        e: this.showJSONEditor,
        g: this.showGifExporter,
        t: this.showTextEditor,
        s: this.showShaderEditor,
        p: this.showPerformanceChart,
        i: this.showLibrary,
        a: this.showAudio,
        w: this.width,
        h: this.height,
        f: this.fps,
        bg: this.backgroundColor,
        mode: this.viewMode,
        fs: this.showFileSettings,
      }),
      /* setState */ (newState) => {
        this.showLottie = !!newState.l;
        this.showJSONEditor = !!newState.e;
        this.showGifExporter = !!newState.g;
        this.showTextEditor = !!newState.t;
        this.showShaderEditor = !!newState.s;
        this.showPerformanceChart = !!newState.p;
        this.showLibrary = !!newState.i;
        this.showAudio = !!newState.a;
        this.width = +newState.w;
        this.height = +newState.h;
        this.fps = +newState.f;
        this.showFileSettings = !!newState.fs;
        this.viewMode =
          newState.mode === 'presentation' ? 'presentation' : 'default';
        this.backgroundColor = String(newState.bg);
        this.render();
      }
    );
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.reflectFromURL();
    window.addEventListener('popstate', this.reflectFromURL);
    this.render();

    // Start a continuous animation loop.
    const drawFrame = () => {
      window.requestAnimationFrame(drawFrame);

      // Elsewhere, the _previousFrameTime is set to null to restart
      // the animation. If null, we assume the user hit re-wind
      // and restart both the Skottie animation and the lottie-web one.
      // This avoids the (small) boot-up lag while we wait for the
      // skottie animation to be parsed and loaded.
      if (!this.previousFrameTime && this.playing) {
        this.previousFrameTime = Date.now();
        this.elapsedTime = 0;
      }
      if (this.playing && this.duration > 0) {
        const currentTime = Date.now();
        this.elapsedTime += (currentTime - this.previousFrameTime) * this.speed;
        this.previousFrameTime = currentTime;
        const progress = this.elapsedTime % this.duration;

        // If we want to have synchronized playing, it's best to force
        // all players to draw the same frame rather than letting them play
        // on their own timeline.
        const normalizedProgress = progress / this.duration;
        this.performanceChart?.start(
          progress,
          this.duration,
          this.state.lottie?.fr || 0
        );
        this.skottiePlayer?.seek(normalizedProgress);
        this.performanceChart?.end();
        this.skottieLibrary?.seek(normalizedProgress);

        // lottie player takes the milliseconds from the beginning of the animation.
        this.lottiePlayer?.goToAndStop(progress);
        this.live?.goToAndStop(progress);
        this.updateScrubber();
        this.updateFrameLabel();
      }
    };

    window.requestAnimationFrame(drawFrame);
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    window.removeEventListener('popstate', this.reflectFromURL);
  }

  attributeChangedCallback(): void {
    this.render();
  }

  private updateAnimation(e: CustomEvent<LottieAnimation>): void {
    this.state.lottie = e.detail;
    this.state.filename = e.detail.metadata?.filename || this.state.filename;
    this.upload();
  }

  private applyTextEdits(e: CustomEvent<TextEditApplyEventDetail>): void {
    const texts = e.detail.texts;
    this.state.lottie = replaceTexts(texts, this.state.lottie!);
    this.skottieLibrary?.replaceTexts(texts);

    this.upload();
  }

  private applyShaderEdits(e: CustomEvent<ShaderEditApplyEventDetail>): void {
    const shaders = e.detail.shaders;
    this.state.lottie = replaceShaders(shaders, this.state.lottie!);
    // TODO(jmbetancourt): support skottieLibrary
    // this.skottieLibrary?.replaceShaders(shaders);

    this.upload();
  }

  private applyAudioSync(e: CustomEvent<AudioStartEventDetail>): void {
    const detail = e.detail;
    this.speed = detail.speed;
    this.previousFrameTime = Date.now();
    this.elapsedTime = 0;
    if (!this.playing) {
      this.playpause();
    }
  }

  private onExportStart(): void {
    if (this.playing) {
      this.playpause();
    }
  }

  private applyEdits(): void {
    if (!this.editor || !this.editorLoaded || !this.hasEdits) {
      return;
    }
    this.state.lottie = this.editor.get();
    this.upload();
  }

  private autoSize(): boolean {
    let changed = false;
    if (!this.width) {
      this.width = this.state.lottie!.w;
      changed = true;
    }
    if (!this.height) {
      this.height = this.state.lottie!.h;
      changed = true;
    }
    // By default, leave FPS at 0, instead of reading them from the lottie,
    // because that will cause it to render as smoothly as possible,
    // which looks better in most cases. If a user gives a negative value
    // for fps (e.g. -1), then we use either what the lottie tells us or
    // as fast as possible.
    if (this.fps < 0) {
      this.fps = this.state.lottie!.fr || 0;
    }
    return changed;
  }

  private skottieFileSelected(e: CustomEvent<SkottieConfigEventDetail>) {
    this.state = e.detail.state;
    this.width = e.detail.width;
    this.height = e.detail.height;
    this.fps = e.detail.fps;
    this.backgroundColor = e.detail.backgroundColor;
    this.autoSize();
    this.stateChanged();
    if (e.detail.fileChanged) {
      this.upload();
    } else {
      this.ui = 'loaded';
      this.render();
      this.initializePlayer();
      // Re-sync all players
      this.rewind();
    }
  }

  private skottieFileSettingsUpdated(
    e: CustomEvent<SkottieFileSettingsEventDetail>
  ) {
    this.width = e.detail.width;
    this.height = e.detail.height;
    this.fps = e.detail.fps;
    this.autoSize();
    this.stateChanged();
    this.render();
    this.initializePlayer();
    // Re-sync all players
    this.rewind();
  }

  private selectionCancelled() {
    this.ui = 'loaded';
    this.render();
    this.initializePlayer();
  }

  private initializePlayer(): Promise<void> {
    return this.skottiePlayer!.initialize({
      width: this.width,
      height: this.height,
      lottie: this.state.lottie!,
      assets: this.state.assets,
      soundMap: this.state.soundMap,
      fps: this.fps,
    }).then(() => {
      this.performanceChart?.reset();
      this.duration = this.skottiePlayer!.duration();
      // If the user has specified a value for FPS, we want to lock the
      // size of the scrubber so it is as discrete as the frame rate.
      if (this.fps) {
        const scrubber = $$<HTMLInputElement>('#scrub', this);
        if (scrubber) {
          // calculate a scaled version of ms per frame as the step size.
          scrubber.step = String(
            (1000 / this.fps) * (SCRUBBER_RANGE / this.duration)
          );
        }
      }
      const frameTotal = $$<HTMLInputElement>('#frameTotal', this);
      if (frameTotal) {
        if (this.state.lottie!.fr) {
          frameTotal.textContent =
            'of ' +
            String(Math.round(this.duration * (this.state.lottie!.fr / 1000)));
        }
      }
    });
  }

  private loadAssetsAndRender(): Promise<void> {
    const toLoad: Promise<LoadedAsset | null>[] = [];

    const lottie = this.state.lottie!;
    let fonts: FontAsset[] = [];
    let assets: LottieAsset[] = [];
    if (lottie.fonts && lottie.fonts.list) {
      fonts = lottie.fonts.list;
    }
    if (lottie.assets && lottie.assets.length) {
      assets = lottie.assets;
    }

    toLoad.push(...this.loadFonts(fonts));
    toLoad.push(...this.loadAssets(assets));

    return Promise.all(toLoad)
      .then((externalAssets: (LoadedAsset | null)[]) => {
        const loadedAssets: Record<string, ArrayBuffer> = {};
        const sounds = new SoundMap();
        for (const asset of externalAssets) {
          if (asset && asset.bytes) {
            loadedAssets[asset.name] = asset.bytes;
          } else if (asset && asset.player) {
            sounds.setPlayer(asset.name, asset.player);
          }
        }

        // check fonts
        fonts.forEach((font: FontAsset) => {
          if (!loadedAssets[font.fName]) {
            console.error(`Could not load font '${font.fName}'.`);
          }
        });

        this.state.assets = loadedAssets;
        this.state.soundMap = sounds;
        this.render();
        return this.initializePlayer().then(() => {
          // Re-sync all players
          this.rewind();
        });
      })
      .catch(() => {
        this.render();
        return this.initializePlayer().then(() => {
          // Re-sync all players
          this.rewind();
        });
      });
  }

  private loadFonts(fonts: FontAsset[]): Promise<LoadedAsset | null>[] {
    const promises: Promise<LoadedAsset | null>[] = [];
    for (const font of fonts) {
      if (!font.fName) {
        continue;
      }

      const fetchFont = (fontURL: string) => {
        promises.push(
          fetch(fontURL).then((resp: Response) => {
            // fetch does not reject on 404
            if (!resp.ok) {
              return null;
            }
            return resp.arrayBuffer().then((buffer: ArrayBuffer) => ({
              name: font.fName,
              bytes: buffer,
            }));
          })
        );
      };

      // We have a mirror of google web fonts with a flattened directory structure which
      // makes them easier to find. Additionally, we can host the full .ttf
      // font, instead of the .woff2 font which is served by Google due to
      // it's smaller size by being a subset based on what glyphs are rendered.
      // Since we don't know all the glyphs we need up front, it's easiest
      // to just get the full font as a .ttf file.
      fetchFont(`${GOOGLE_WEB_FONTS_HOST}/${font.fName}.ttf`);

      // Also try using uploaded assets.
      // We may end up with two different blobs for the same font name, in which case
      // the user-provided one takes precedence.
      fetchFont(`${this.assetsPath}/${this.hash}/${font.fName}.ttf`);
    }
    return promises;
  }

  private loadAssets(assets: LottieAsset[]): Promise<LoadedAsset | null>[] {
    const promises: Promise<LoadedAsset | null>[] = [];
    for (const asset of assets) {
      if (asset.id.startsWith('audio_')) {
        // Howler handles our audio assets, they don't provide a promise when making a new Howl.
        // We push the audio asset as is and hope that it loads before playback starts.
        const inline =
          asset.p && asset.p.startsWith && asset.p.startsWith('data:');
        if (inline) {
          promises.push(
            Promise.resolve({
              name: asset.id,
              player: new AudioPlayer(asset.p),
            })
          );
        } else {
          promises.push(
            Promise.resolve({
              name: asset.id,
              player: new AudioPlayer(
                `${this.assetsPath}/${this.hash}/${asset.p}`
              ),
            })
          );
        }
      } else {
        // asset.p is the filename, if it's an image.
        // Don't try to load inline/dataURI images.
        const should_load =
          asset.p && asset.p.startsWith && !asset.p.startsWith('data:');
        if (should_load) {
          promises.push(
            fetch(`${this.assetsPath}/${this.hash}/${asset.p}`).then(
              (resp: Response) => {
                // fetch does not reject on 404
                if (!resp.ok) {
                  console.error(
                    `Could not load ${asset.p}: status ${resp.status}`
                  );
                  return null;
                }
                return resp.arrayBuffer().then((buffer) => ({
                  name: asset.p,
                  bytes: buffer,
                }));
              }
            )
          );
        }
      }
    }
    return promises;
  }

  private playpause(): void {
    const audioManager = $$<SkottieAudioSk>('skottie-audio-sk');
    if (this.playing) {
      this.lottiePlayer?.pause();
      this.live?.pause();
      this.state.soundMap?.pause();
      $$<HTMLElement>('#playpause-pause')!.style.display = 'none';
      $$<HTMLElement>('#playpause-play')!.style.display = 'inherit';
      audioManager?.pause();
    } else {
      this.lottiePlayer?.play();
      this.live?.play();
      this.previousFrameTime = Date.now();
      // There is no need call a soundMap.play() function here.
      // Skottie invokes the play by calling seek on the needed audio track.
      $$<HTMLElement>('#playpause-pause')!.style.display = 'inherit';
      $$<HTMLElement>('#playpause-play')!.style.display = 'none';
      audioManager?.resume();
    }
    this.playing = !this.playing;
  }

  private recoverFromError(msg: string): void {
    errorMessage(msg);
    console.error(msg);
    window.history.pushState(null, '', '/');
    this.ui = 'dialog';
    this.render();
  }

  private reflectFromURL(): void {
    // Check URL.
    const match = window.location.pathname.match(/\/([a-zA-Z0-9]+)/);
    if (!match) {
      this.hash = DEFAULT_LOTTIE_FILE;
    } else {
      this.hash = match[1];
    }
    this.ui = 'loading';
    this.render();
    // Run this on the next micro-task to allow mocks to be set up if needed.
    setTimeout(() => {
      fetch(`/_/j/${this.hash}`, {
        credentials: 'include',
      })
        .then(jsonOrThrow)
        .then((json) => {
          // remove legacy fields from state, if they are there.
          delete json.width;
          delete json.height;
          delete json.fps;
          this.state = json;

          if (this.autoSize()) {
            this.stateChanged();
          }
          this.ui = 'loaded';
          this.loadAssetsAndRender().then(() => {
            console.log('loaded');
            this.dispatchEvent(
              new CustomEvent('initial-animation-loaded', { bubbles: true })
            );
          });
        })
        .catch((msg) => this.recoverFromError(msg));
    });
  }

  private render(): void {
    if (this.downloadURL) {
      URL.revokeObjectURL(this.downloadURL);
    }
    this.downloadURL = URL.createObjectURL(
      new Blob([JSON.stringify(this.state.lottie)])
    );
    super._render();

    this.skottiePlayer = $$<SkottiePlayerSk>('skottie-player-sk', this);
    this.performanceChart = $$<SkottiePerformanceSk>(
      'skottie-performance-sk',
      this
    );
    this.skottieLibrary = $$<SkottieLibrarySk>('skottie-library-sk', this);

    const skottieGifExporter = $$<SkottieGifExporterSk>(
      'skottie-gif-exporter-sk',
      this
    );
    if (skottieGifExporter && this.skottiePlayer) {
      skottieGifExporter.player = this.skottiePlayer;
    }

    if (this.ui === 'loaded') {
      if (this.state.soundMap && this.state.soundMap.map.size > 0) {
        this.hideVolumeSlider(false);
        // Stop any audio assets that start playing on frame 0
        // Pause the playback to force a user gesture to resume the AudioContext
        if (this.playing) {
          this.playpause();
          this.rewind();
        }
        this.state.soundMap.stop();
      } else {
        this.hideVolumeSlider(true);
      }
      try {
        this.renderLottieWeb();
        this.renderJSONEditor();
        this.renderTextEditor();
        this.renderShaderEditor();
        this.renderAudioManager();
      } catch (e) {
        console.warn('caught error while rendering third party code', e);
      }
    }
  }

  private renderAudioManager(): void {
    if (this.showAudio) {
      const audioManager = $$<SkottieAudioSk>('skottie-audio-sk', this);
      if (audioManager) {
        audioManager.animation = this.state.lottie!;
      }
    }
  }

  private renderTextEditor(): void {
    if (this.showTextEditor) {
      const textEditor = $$<SkottieTextEditorSk>(
        'skottie-text-editor-sk',
        this
      );
      if (textEditor) {
        textEditor.animation = this.state.lottie!;
      }
    }
  }

  private renderShaderEditor(): void {
    if (this.showShaderEditor) {
      const shaderEditor = $$<ShaderEditorSk>('skottie-shader-editor-sk', this);
      if (shaderEditor) {
        shaderEditor.animation = this.state.lottie!;
      }
    }
  }

  private renderJSONEditor(): void {
    if (!this.showJSONEditor) {
      this.editorLoaded = false;
      this.editor = null;
      return;
    }
    const editorContainer = $$<HTMLDivElement>('#json_editor')!;
    // See https://github.com/josdejong/jsoneditor/blob/master/docs/api.md
    // for documentation on this editor.
    const editorOptions = {
      // Use original key order (this preserves related fields locality).
      sortObjectKeys: false,
      // There are sometimes a few onChange events that happen
      // during the initial .set(), so we have a safety variable
      // _editorLoaded to prevent a bunch of recursion
      onChange: () => {
        if (!this.editorLoaded) {
          return;
        }
        this.hasEdits = true;
        this.render();
      },
    };

    if (!this.editor) {
      this.editorLoaded = false;
      editorContainer.innerHTML = '';
      this.editor = new JSONEditor(editorContainer, editorOptions);
    }
    if (!this.hasEdits) {
      this.editorLoaded = false;
      // Only set the JSON when it is loaded, either because it's
      // the first time we got it from the server or because the user
      // hit applyEdits.
      this.editor.set(this.state.lottie);
    }
    // We are now pretty confident that the onChange events will only be
    // when the user modifies the JSON.
    this.editorLoaded = true;
  }

  private renderLottieWeb(): void {
    if (!this.showLottie) {
      return;
    }
    // Don't re-start the animation while the user edits.
    if (!this.hasEdits) {
      $$<HTMLDivElement>('#container')!.innerHTML = '';
      this.lottiePlayer = LottiePlayer.loadAnimation({
        container: $$('#container')!,
        renderer: 'svg',
        loop: true,
        autoplay: this.playing,
        assetsPath: `${this.assetsPath}/${this.hash}/`,
        // Apparently the lottie player modifies the data as it runs?
        animationData: JSON.parse(
          JSON.stringify(this.state.lottie)
        ) as LottieAnimation,
        rendererSettings: {
          preserveAspectRatio: 'xMidYMid meet',
        },
      });
      this.live = null;
    } else {
      // we have edits, update the live preview version.
      // It will re-start from the very beginning, but the user can
      // hit "rewind" to re-sync them.
      $$<HTMLDivElement>('#live')!.innerHTML = '';
      this.live = LottiePlayer.loadAnimation({
        container: $$('#live')!,
        renderer: 'svg',
        loop: true,
        autoplay: this.playing,
        assetsPath: `${this.assetsPath}/${this.hash}/`,
        // Apparently the lottie player modifies the data as it runs?
        animationData: JSON.parse(
          JSON.stringify(this.editor!.get())
        ) as LottieAnimation,
        rendererSettings: {
          preserveAspectRatio: 'xMidYMid meet',
        },
      });
    }
  }

  // This fires every time the user moves the scrub slider.
  private onScrub(e: Event): void {
    if (!this.scrubbing) {
      // Pause the animation while dragging the slider.
      this.playingOnStartOfScrub = this.playing;
      if (this.playing) {
        this.playpause();
      }
      this.scrubbing = true;
    }
    const scrubber = (e.target as HTMLInputElement)!;
    const seek = +scrubber.value / SCRUBBER_RANGE;
    this.seek(seek);
    this.updateFrameLabel();
  }

  // This fires when the user releases the scrub slider.
  private onScrubEnd(): void {
    if (this.playingOnStartOfScrub) {
      this.playpause();
    }
    this.scrubbing = false;
  }

  private onFrameFocus(): void {
    if (this.playing) {
      this.playpause();
    }
  }

  private onFrameChange(e: Event): void {
    if (this.playing) {
      this.playpause();
    }
    const frameInput = $$<HTMLInputElement>('#frameInput', this);
    if (frameInput) {
      const frame = +frameInput.value;
      this.seekFrame(frame);
    }
  }

  private onChartClick(e: Event): void {
    const chart = $$<SkottiePerformanceSk>('#chart', this);
    const frame: number | undefined = chart?.getClickedFrame(e);
    if (frame !== undefined && frame !== -1) {
      if (this.playing) {
        this.playpause();
      }
      const frameInput = $$<HTMLInputElement>('#frameInput', this);
      if (frameInput) frameInput.value = String(frame);
      this.seekFrame(frame);
    }
  }

  private seekFrame(frame: number): void {
    if (frame > 0 && frame < this.duration) {
      let seek = 0;
      if (this.state.lottie?.fr) {
        seek = ((frame / this.state.lottie.fr) * 1000) / this.duration;
      }
      this.seek(seek);
      this.updateScrubber();
    }
  }

  private updateScrubber(): void {
    const scrubber = $$<HTMLInputElement>('#scrub', this);
    if (scrubber) {
      // Scale from time to the arbitrary scrubber range.
      const progress = this.elapsedTime % this.duration;
      scrubber.value = String((SCRUBBER_RANGE * progress) / this.duration);
    }
  }

  private updateFrameLabel(): void {
    const frameLabel = $$<HTMLInputElement>('#frameInput', this);
    if (frameLabel) {
      const progress = this.elapsedTime % this.duration;
      if (this.state.lottie!.fr) {
        frameLabel.value = String(
          Math.round(progress * (this.state.lottie!.fr / 1000))
        );
      }
    }
  }

  private seek(t: number): void {
    // catch case where t = 1
    t = Math.min(t, 0.9999);
    this.elapsedTime = t * this.duration;
    this.live?.goToAndStop(t);
    this.lottiePlayer?.goToAndStop(t * this.duration);
    this.skottiePlayer?.seek(t);
    this.skottieLibrary?.seek(t);
  }

  private onVolumeChange(e: Event): void {
    const scrubber = (e.target as HTMLInputElement)!;
    this.state.soundMap?.setVolume(+scrubber.value);
  }

  private rewind(): void {
    // Handle rewinding when paused.
    if (!this.playing) {
      this.skottiePlayer!.seek(0);
      this.skottieLibrary?.seek(0);
      this.previousFrameTime = 0;
      this.live?.goToAndStop(0);
      this.lottiePlayer?.goToAndStop(0);
      const scrubber = $$<HTMLInputElement>('#scrub', this);
      if (scrubber) {
        scrubber.value = '0';
      }
    } else {
      this.live?.goToAndPlay(0);
      this.lottiePlayer?.goToAndPlay(0);
      this.previousFrameTime = 0;
      const audioManager = $$<SkottieAudioSk>('skottie-audio-sk', this);
      audioManager?.rewind();
    }
  }

  private startEdit(): void {
    this.ui = 'dialog';
    this.render();
  }

  private toggleEditor(e: Event): void {
    // avoid double toggles
    e.preventDefault();
    this.showTextEditor = false;
    this.showJSONEditor = !this.showJSONEditor;
    this.stateChanged();
    this.render();
  }

  private toggleGifExporter(e: Event): void {
    // avoid double toggles
    e.preventDefault();
    this.showGifExporter = !this.showGifExporter;
    this.stateChanged();
    this.render();
  }

  private exportSelectHandler(e: CustomEvent<DropdownSelectEvent>): void {
    if (!this.skottiePlayer) {
      return;
    }
    const exportManager = $$<SkottieExporterSk>('skottie-exporter-sk');
    exportManager?.export(e.detail.value as ExportType, this.skottiePlayer);
  }

  private togglePerformanceChart(e: Event): void {
    // avoid double toggles
    e.preventDefault();
    this.showPerformanceChart = !this.showPerformanceChart;
    this.stateChanged();
    this.render();
  }

  private toggleTextEditor(open: boolean): void {
    this.showJSONEditor = false;
    this.showTextEditor = open;
    this.stateChanged();
    this.render();
  }

  private toggleShaderEditor(open: boolean): void {
    this.showJSONEditor = false;
    this.showShaderEditor = open;
    this.stateChanged();
    this.render();
  }

  private toggleLibrary(open: boolean): void {
    this.showLibrary = open;
    this.stateChanged();
    this.render();
  }

  private toggleAudio(open: boolean): void {
    this.showAudio = open;
    this.stateChanged();
    this.render();
  }

  private toggleFileSettings(open: boolean): void {
    this.showFileSettings = open;
    this.stateChanged();
    this.render();
  }

  private toggleLottie(e: Event): void {
    // avoid double toggles
    e.preventDefault();
    this.showLottie = !this.showLottie;
    this.stateChanged();
    this.render();
  }

  private hideVolumeSlider(v: boolean) {
    const collapse = $$<CollapseSk>('#volume', this);
    if (collapse) {
      collapse.closed = v;
    }
  }

  private upload(): void {
    // POST the JSON to /_/upload
    this.hash = '';
    this.hasEdits = false;
    this.editorLoaded = false;
    this.editor = null;
    // Clean up the old animation and other wasm objects
    this.render();
    fetch('/_/upload', {
      credentials: 'include',
      body: JSON.stringify(this.state),
      headers: {
        'Content-Type': 'application/json',
      },
      method: 'POST',
    })
      .then(jsonOrThrow)
      .then((json) => {
        // Should return with the hash and the lottie file
        this.ui = 'loaded';
        this.hash = json.hash;
        this.state.lottie = json.lottie;
        window.history.pushState(null, '', `/${this.hash}`);
        this.stateChanged();
        if (this.state.assetsZip) {
          this.loadAssetsAndRender();
        }
        this.render();
      })
      .catch((msg) => this.recoverFromError(msg));

    if (!this.state.assetsZip) {
      this.ui = 'loaded';
      // Start drawing right away, no need to wait for
      // the JSON to make a round-trip to the server, since there
      // are no assets that we need to unzip server-side.
      // We still need to check for things like webfonts.
      this.render();
      this.loadAssetsAndRender();
    } else {
      // We have to wait for the server to process the zip file.
      this.ui = 'loading';
      this.render();
    }
  }

  overrideAssetsPathForTesting(p: string): void {
    this.assetsPath = p;
  }
}

define('skottie-sk', SkottieSk);
