const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  await page.goto('http://localhost:8318');
  await page.waitForTimeout(1000);

  // Add dark class to html element (class-based dark mode)
  await page.evaluate(() => {
    document.documentElement.classList.add('dark');
  });
  await page.waitForTimeout(1000);

  await page.click('text=Accounts');
  await page.waitForTimeout(1000);

  await page.click('text=Add Account');
  await page.waitForTimeout(1000);

  await page.screenshot({ path: '/workspace/llm-mux/ui/dialog-screenshot.png', fullPage: true });

  const dialogStyles = await page.evaluate(() => {
    const dialog = document.querySelector('[role="dialog"]');
    if (!dialog) return { error: 'No dialog found' };

    const style = window.getComputedStyle(dialog);
    return {
      backgroundColor: style.backgroundColor,
      color: style.color,
      classes: dialog.className
    };
  });

  console.log('Dialog styles:', JSON.stringify(dialogStyles, null, 2));

  const buttonStyles = await page.evaluate(() => {
    const buttons = document.querySelectorAll('[role="dialog"] button');
    return Array.from(buttons).slice(0, 3).map(btn => {
      const style = window.getComputedStyle(btn);
      return {
        text: btn.textContent?.substring(0, 30),
        backgroundColor: style.backgroundColor,
        color: style.color
      };
    });
  });

  console.log('Button styles:', JSON.stringify(buttonStyles, null, 2));

  await browser.close();
})();
