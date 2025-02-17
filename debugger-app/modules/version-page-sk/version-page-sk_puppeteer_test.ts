import { expect } from 'chai';
import {
  inBazel,
  loadCachedTestBed,
  takeScreenshot,
  TestBed,
} from '../../../puppeteer-tests/util';

describe('version-page-sk', () => {
  let testBed: TestBed;
  before(async () => {
    testBed = await loadCachedTestBed();
  });

  beforeEach(async () => {
    // Remove the /dist/ below for //infra-sk elements.
    await testBed.page.goto(
      inBazel()
        ? testBed.baseUrl
        : `${testBed.baseUrl}/dist/version-page-sk.html`
    );
    await testBed.page.setViewport({ width: 400, height: 550 });
  });

  it('should render the demo page (smoke test)', async () => {
    expect(await testBed.page.$$('version-page-sk')).to.have.length(1);
  });

  describe('screenshots', () => {
    it('shows the default view', async () => {
      await takeScreenshot(testBed.page, 'debugger-app', 'version-page-sk');
    });
  });
});
