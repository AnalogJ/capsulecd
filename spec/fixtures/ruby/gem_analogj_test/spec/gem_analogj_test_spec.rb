require 'spec_helper'

describe 'GemTest' do
  it 'has a version number' do
    expect(GemTest::VERSION).not_to be nil
  end

  it 'does something useful' do
    expect(true).to eq(true)
  end
end
